// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	kubemarkers "sigs.k8s.io/controller-tools/pkg/markers"

	"github.com/solo-io/protoc-gen-openapi/pkg/markers"
	"github.com/solo-io/protoc-gen-openapi/pkg/protomodel"
)

var descriptionExclusionMarkers = []string{"$hide_from_docs", "$hide", "@exclude"}

// Some special types with predefined schemas.
// This is to catch cases where solo apis contain recursive definitions
// Normally these would result in stack-overflow errors when generating the open api schema
// The imperfect solution, is to just generate an empty object for these types
var specialSoloTypes = map[string]openapi3.Schema{
	"core.solo.io.Metadata": {
		Type: &openapi3.Types{openapi3.TypeObject},
	},
	"google.protobuf.ListValue": *openapi3.NewArraySchema().WithItems(openapi3.NewObjectSchema()),
	"google.protobuf.Struct": {
		Type:       &openapi3.Types{openapi3.TypeObject},
		Properties: make(map[string]*openapi3.SchemaRef),
		Extensions: map[string]interface{}{
			"x-kubernetes-preserve-unknown-fields": true,
		},
	},
	"google.protobuf.Any": {
		Type:       &openapi3.Types{openapi3.TypeObject},
		Properties: make(map[string]*openapi3.SchemaRef),
		Extensions: map[string]interface{}{
			"x-kubernetes-preserve-unknown-fields": true,
		},
	},
	"google.protobuf.Value": {
		Properties: make(map[string]*openapi3.SchemaRef),
		Extensions: map[string]interface{}{
			"x-kubernetes-preserve-unknown-fields": true,
		},
	},
	"google.protobuf.BoolValue":   *openapi3.NewBoolSchema().WithNullable(),
	"google.protobuf.StringValue": *openapi3.NewStringSchema().WithNullable(),
	"google.protobuf.DoubleValue": *openapi3.NewFloat64Schema().WithNullable(),
	"google.protobuf.Int32Value":  *openapi3.NewIntegerSchema().WithNullable().WithMin(math.MinInt32).WithMax(math.MaxInt32),
	"google.protobuf.Int64Value":  *openapi3.NewIntegerSchema().WithNullable().WithMin(math.MinInt64).WithMax(math.MaxInt64),
	"google.protobuf.UInt32Value": *openapi3.NewIntegerSchema().WithNullable().WithMin(0).WithMax(math.MaxUint32),
	"google.protobuf.UInt64Value": *openapi3.NewIntegerSchema().WithNullable().WithMin(0).WithMax(math.MaxUint64),
	"google.protobuf.FloatValue":  *openapi3.NewFloat64Schema().WithNullable(),
	"google.protobuf.Duration":    *openapi3.NewStringSchema(),
	"google.protobuf.Empty":       *openapi3.NewObjectSchema().WithMaxProperties(0),
	"google.protobuf.Timestamp":   *openapi3.NewStringSchema().WithFormat("date-time"),
}

type openapiGenerator struct {
	buffer     bytes.Buffer
	model      *protomodel.Model
	perFile    bool
	singleFile bool
	yaml       bool
	useRef     bool

	// transient state as individual files are processed
	currentPackage             *protomodel.PackageDescriptor
	currentFrontMatterProvider *protomodel.FileDescriptor

	messages map[string]*protomodel.MessageDescriptor

	// @solo.io customizations to limit length of generated descriptions
	descriptionConfiguration *DescriptionConfiguration

	// @solo.io customization to support enum validation schemas with int or string values
	// we need to support this since some controllers marshal enums as integers and others as strings
	enumAsIntOrString bool

	// @solo.io customizations to define schemas for certain messages
	customSchemasByMessageName map[string]openapi3.Schema

	// If set to true, OpenAPI schema will include schema to emulate behavior of protobuf oneof fields
	protoOneof bool

	// If set to true, native OpenAPI integer scehmas will be used for integer types instead of Solo wrappers
	// that add Kubernetes extension headers to the schema to treat int as strings.
	intNative bool

	markerRegistry *markers.Registry

	// If set to true, kubebuilder markers and validations such as PreserveUnknownFields, MinItems, default, and all CEL rules will be omitted from the OpenAPI schema.
	// The Type and Required markers will be maintained.
	disableKubeMarkers bool

	// when set, this list of substrings will be used to identify kubebuilder markers to ignore. When multiple are
	// supplied, this will function as a logical OR i.e. any rule which contains a provided substring will be ignored
	ignoredKubeMarkerSubstrings []string
}

type DescriptionConfiguration struct {
	// Whether or not to include a description in the generated open api schema
	IncludeDescriptionInSchema bool

	// Whether or not the description for properties should be allowed to span multiple lines
	MultilineDescription bool
}

func newOpenAPIGenerator(
	model *protomodel.Model,
	perFile bool,
	singleFile bool,
	yaml bool,
	useRef bool,
	descriptionConfiguration *DescriptionConfiguration,
	enumAsIntOrString bool,
	messagesWithEmptySchema []string,
	protoOneof bool,
	intNative bool,
	disableKubeMarkers bool,
	ignoredKubeMarkers []string,
) *openapiGenerator {
	mRegistry, err := markers.NewRegistry()
	if err != nil {
		log.Panicf("error initializing marker registry: %v", err)
	}
	return &openapiGenerator{
		model:                       model,
		perFile:                     perFile,
		singleFile:                  singleFile,
		yaml:                        yaml,
		useRef:                      useRef,
		descriptionConfiguration:    descriptionConfiguration,
		enumAsIntOrString:           enumAsIntOrString,
		customSchemasByMessageName:  buildCustomSchemasByMessageName(messagesWithEmptySchema),
		protoOneof:                  protoOneof,
		intNative:                   intNative,
		markerRegistry:              mRegistry,
		disableKubeMarkers:          disableKubeMarkers,
		ignoredKubeMarkerSubstrings: ignoredKubeMarkers,
	}
}

// buildCustomSchemasByMessageName name returns a mapping of message name to a pre-defined openapi schema
// It includes:
//  1. `specialSoloTypes`, a set of pre-defined schemas
//  2. `messagesWithEmptySchema`, a list of messages that are injected at runtime that should contain
//     and empty schema which accepts and preserves all fields
func buildCustomSchemasByMessageName(messagesWithEmptySchema []string) map[string]openapi3.Schema {
	schemasByMessageName := make(map[string]openapi3.Schema)

	// Initialize the hard-coded values
	for name, schema := range specialSoloTypes {
		schemasByMessageName[name] = schema
	}

	// Add the messages that were injected at runtime
	for _, messageName := range messagesWithEmptySchema {
		emptyMessage := openapi3.Schema{
			Type:       &openapi3.Types{openapi3.TypeObject},
			Properties: make(map[string]*openapi3.SchemaRef),
			Extensions: map[string]interface{}{
				"x-kubernetes-preserve-unknown-fields": true,
			},
		}
		schemasByMessageName[messageName] = emptyMessage
	}

	return schemasByMessageName
}

func (g *openapiGenerator) generateOutput(filesToGen map[*protomodel.FileDescriptor]bool) (*pluginpb.CodeGeneratorResponse, error) {
	response := pluginpb.CodeGeneratorResponse{}

	if g.singleFile {
		g.generateSingleFileOutput(filesToGen, &response)
	} else {
		for _, pkg := range g.model.Packages {
			g.currentPackage = pkg

			// anything to output for this package?
			count := 0
			for _, file := range pkg.Files {
				if _, ok := filesToGen[file]; ok {
					count++
				}
			}

			if count > 0 {
				if g.perFile {
					g.generatePerFileOutput(filesToGen, pkg, &response)
				} else {
					g.generatePerPackageOutput(filesToGen, pkg, &response)
				}
			}
		}
	}

	return &response, nil
}

func (g *openapiGenerator) getFileContents(file *protomodel.FileDescriptor,
	messages map[string]*protomodel.MessageDescriptor,
	enums map[string]*protomodel.EnumDescriptor,
	services map[string]*protomodel.ServiceDescriptor,
) {
	for _, m := range file.AllMessages {
		messages[g.relativeName(m)] = m
	}

	for _, e := range file.AllEnums {
		enums[g.relativeName(e)] = e
	}

	for _, s := range file.Services {
		services[g.relativeName(s)] = s
	}
}

func (g *openapiGenerator) generatePerFileOutput(filesToGen map[*protomodel.FileDescriptor]bool, pkg *protomodel.PackageDescriptor,
	response *pluginpb.CodeGeneratorResponse,
) {
	for _, file := range pkg.Files {
		if _, ok := filesToGen[file]; ok {
			g.currentFrontMatterProvider = file
			messages := make(map[string]*protomodel.MessageDescriptor)
			enums := make(map[string]*protomodel.EnumDescriptor)
			services := make(map[string]*protomodel.ServiceDescriptor)

			g.getFileContents(file, messages, enums, services)
			filename := path.Base(file.GetName())
			extension := path.Ext(filename)
			name := filename[0 : len(filename)-len(extension)]

			rf := g.generateFile(name, file, messages, enums, services)
			response.File = append(response.File, &rf)
		}
	}
}

func (g *openapiGenerator) generateSingleFileOutput(filesToGen map[*protomodel.FileDescriptor]bool, response *pluginpb.CodeGeneratorResponse) {
	messages := make(map[string]*protomodel.MessageDescriptor)
	enums := make(map[string]*protomodel.EnumDescriptor)
	services := make(map[string]*protomodel.ServiceDescriptor)

	for file, ok := range filesToGen {
		if ok {
			g.getFileContents(file, messages, enums, services)
		}
	}

	rf := g.generateFile("openapiv3", &protomodel.FileDescriptor{}, messages, enums, services)
	response.File = []*pluginpb.CodeGeneratorResponse_File{&rf}
}

func (g *openapiGenerator) generatePerPackageOutput(filesToGen map[*protomodel.FileDescriptor]bool, pkg *protomodel.PackageDescriptor,
	response *pluginpb.CodeGeneratorResponse,
) {
	// We need to produce a file for this package.

	// Decide which types need to be included in the generated file.
	// This will be all the types in the fileToGen input files, along with any
	// dependent types which are located in packages that don't have
	// a known location on the web.
	messages := make(map[string]*protomodel.MessageDescriptor)
	enums := make(map[string]*protomodel.EnumDescriptor)
	services := make(map[string]*protomodel.ServiceDescriptor)

	g.currentFrontMatterProvider = pkg.FileDesc()

	for _, file := range pkg.Files {
		if _, ok := filesToGen[file]; ok {
			g.getFileContents(file, messages, enums, services)
		}
	}

	rf := g.generateFile(pkg.Name, pkg.FileDesc(), messages, enums, services)
	response.File = append(response.File, &rf)
}

// Generate an OpenAPI spec for a collection of cross-linked files.
func (g *openapiGenerator) generateFile(name string,
	pkg *protomodel.FileDescriptor,
	messages map[string]*protomodel.MessageDescriptor,
	enums map[string]*protomodel.EnumDescriptor,
	_ map[string]*protomodel.ServiceDescriptor,
) pluginpb.CodeGeneratorResponse_File {
	g.messages = messages

	allSchemas := make(map[string]*openapi3.SchemaRef)

	for _, message := range messages {
		// we generate the top-level messages here and the nested messages are generated
		// inside each top-level message.
		if message.Parent == nil {
			g.generateMessage(message, allSchemas)
		}
	}

	for _, enum := range enums {
		// when there is no parent to the enum.
		if len(enum.QualifiedName()) == 1 {
			g.generateEnum(enum, allSchemas)
		}
	}

	var version string
	var description string
	// only get the API version when generate per package or per file,
	// as we cannot guarantee all protos in the input are the same version.
	if !g.singleFile {
		if g.currentFrontMatterProvider != nil && g.currentFrontMatterProvider.Matter.Description != "" {
			description = g.currentFrontMatterProvider.Matter.Description
		} else if pd := g.generateDescription(g.currentPackage); pd != "" {
			description = pd
		} else {
			description = "OpenAPI Spec for Solo APIs."
		}
		// derive the API version from the package name
		// which is a convention for Istio APIs.
		var p string
		if pkg != nil {
			p = pkg.GetPackage()
		} else {
			p = name
		}
		s := strings.Split(p, ".")
		version = s[len(s)-1]
	} else {
		description = "OpenAPI Spec for Solo APIs."
	}

	c := openapi3.NewComponents()
	c.Schemas = allSchemas
	// add the openapi object required by the spec.
	o := openapi3.T{
		OpenAPI: "3.0.1",
		Info: &openapi3.Info{
			Title:   description,
			Version: version,
		},
		Components: &c,
	}

	g.buffer.Reset()
	var filename *string
	if g.yaml {
		b, err := yaml.Marshal(o)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to marshall the output of %v to yaml", name)
		}
		filename = proto.String(name + ".yaml")
		g.buffer.Write(b)
	} else {
		b, err := json.MarshalIndent(o, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to marshall the output of %v to json", name)
		}
		filename = proto.String(name + ".json")
		g.buffer.Write(b)
	}

	return pluginpb.CodeGeneratorResponse_File{
		Name:    filename,
		Content: proto.String(g.buffer.String()),
	}
}

func (g *openapiGenerator) generateMessage(message *protomodel.MessageDescriptor, allSchemas map[string]*openapi3.SchemaRef) {
	if o := g.generateMessageSchema(message); o != nil {
		allSchemas[g.absoluteName(message)] = o.NewRef()
	}
}

func (g *openapiGenerator) generateSoloMessageSchema(message *protomodel.MessageDescriptor, customSchema *openapi3.Schema) *openapi3.Schema {
	o := customSchema
	o.Description = g.generateDescription(message)

	return o
}

func (g *openapiGenerator) generateSoloInt64Schema() *openapi3.Schema {
	schema := openapi3.NewInt64Schema()
	schema.Extensions = map[string]interface{}{
		"x-kubernetes-int-or-string": true,
	}

	return schema
}

func (g *openapiGenerator) generateMessageSchema(message *protomodel.MessageDescriptor) *openapi3.Schema {
	// skip MapEntry message because we handle map using the map's repeated field.
	if message.GetOptions().GetMapEntry() {
		return nil
	}
	o := openapi3.NewObjectSchema()
	o.Description = g.generateDescription(message)
	msgRules := g.validationRules(message)
	g.mustApplyRulesToSchema(msgRules, o, markers.TargetType)

	oneOfFields := make(map[int32][]string)
	var requiredFields []string
		// Handle proto3 optional fields - they should not be required
	for _, field := range message.Fields {
		if field.Proto3Optional != nil && *field.Proto3Optional {
		repeated := field.IsRepeated()
			// Proto3 optional fields are never required
		fieldName := g.fieldName(field)
		} else if g.markerRegistry.IsRequired(fieldRules) {
		fieldDesc := g.generateDescription(field)
		fieldRules := g.validationRules(field)

		// If the field is a oneof, we need to add the oneof property to the schema
		if field.OneofIndex != nil {
			idx := *field.OneofIndex
			oneOfFields[idx] = append(oneOfFields[idx], fieldName)
		}

			// Field is required by marker
			requiredFields = append(requiredFields, fieldName)
		}
		}

		schemaType := g.markerRegistry.GetSchemaType(fieldRules, markers.TargetField)
		if schemaType != "" {
			tmp := getSoloSchemaForMarkerType(schemaType)
			schema := getSchemaIfRepeated(&tmp, repeated)
			schema.Description = fieldDesc
			g.mustApplyRulesToSchema(fieldRules, schema, markers.TargetField)
			o.WithProperty(fieldName, schema)
			continue
		}

		sr := g.fieldTypeRef(field)
		g.mustApplyRulesToSchema(fieldRules, sr.Value, markers.TargetField)
		o.WithProperty(fieldName, sr.Value)
	}

	if len(requiredFields) > 0 {
		o.Required = requiredFields
	}

	if g.protoOneof {
		// Add protobuf oneof schema for this message
		oneOfs := make([][]*openapi3.Schema, len(oneOfFields))
		for idx := range oneOfFields {
			// oneOfSchemas is a collection (not and required schemas) that should be assigned to the schemas's oneOf field
			oneOfSchemas := newProtoOneOfSchema(oneOfFields[idx]...)
			oneOfs[idx] = append(oneOfs[idx], oneOfSchemas...)
		}

		switch len(oneOfs) {
		case 0:
			// no oneof fields
		case 1:
			o.OneOf = getSchemaRefs(oneOfs[0]...)
		default:
			// Wrap collected OneOf refs with AllOf schema
			for _, schemas := range oneOfs {
				oneOfRef := openapi3.NewOneOfSchema(schemas...)
				o.AllOf = append(o.AllOf, oneOfRef.NewRef())
			}
		}
	}

	return o
}

func getSoloSchemaForMarkerType(t markers.Type) openapi3.Schema {
	switch t {
	case markers.TypeObject:
		return specialSoloTypes["google.protobuf.Struct"]
	case markers.TypeValue:
		return specialSoloTypes["google.protobuf.Value"]
	default:
		log.Panicf("unexpected schema type %v", t)
		return openapi3.Schema{}
	}
}

func getSchemaRefs(schemas ...*openapi3.Schema) openapi3.SchemaRefs {
	var refs openapi3.SchemaRefs
	for _, schema := range schemas {
		refs = append(refs, schema.NewRef())
	}
	return refs
}

func getSchemaIfRepeated(schema *openapi3.Schema, repeated bool) *openapi3.Schema {
	if repeated {
		schema = openapi3.NewArraySchema().WithItems(schema)
	}
	return schema
}

// newProtoOneOfSchema returns a schema that can be used to represent a collection of fields
// that must be encoded as a oneOf in OpenAPI.
// For e.g., if the fields x and y are a part of a proto oneof, then they can be represented as
// follows, such that only one of x or y is required and specifying neither is also acceptable.
//
//	{
//		"not": {
//			"anyOf": [
//				{
//					"required": [
//						"x"
//					]
//				},
//				{
//					"required": [
//						"y"
//					]
//				}
//			]
//		}
//	},
//	{
//		"required": [
//			"x"
//		]
//	},
//	{
//		"required": [
//			"y"
//		]
//	}
func newProtoOneOfSchema(fields ...string) []*openapi3.Schema {
	fieldSchemas := make([]*openapi3.Schema, len(fields))
	for i, field := range fields {
		schema := openapi3.NewSchema()
		schema.Required = []string{field}
		fieldSchemas[i] = schema
	}
	// convert fieldSchema to a oneOf schema
	anyOfSchema := openapi3.NewAnyOfSchema(fieldSchemas...)
	notAnyOfSchema := openapi3.NewSchema()
	notAnyOfSchema.Not = anyOfSchema.NewRef()

	allOneOfSchemas := make([]*openapi3.Schema, len(fieldSchemas)+1)
	allOneOfSchemas[0] = notAnyOfSchema
	for i, fieldSchema := range fieldSchemas {
		allOneOfSchemas[i+1] = fieldSchema
	}
	return allOneOfSchemas
}

func (g *openapiGenerator) generateEnum(enum *protomodel.EnumDescriptor, allSchemas map[string]*openapi3.SchemaRef) {
	o := g.generateEnumSchema(enum)
	allSchemas[g.absoluteName(enum)] = o.NewRef()
}

func (g *openapiGenerator) generateEnumSchema(enum *protomodel.EnumDescriptor) *openapi3.Schema {
	/**
	  The out of the box solution created an enum like:
	  	enum:
	  	- - option_a
	  	  - option_b
	  	  - option_c

	  Instead, what we want is:
	  	enum:
	  	- option_a
	  	- option_b
	  	- option_c
	*/
	o := openapi3.NewStringSchema()
	o.Description = g.generateDescription(enum)

	// If the schema should be int or string, mark it as such
	if g.enumAsIntOrString {
		o.Extensions = map[string]interface{}{
			"x-kubernetes-int-or-string": true,
		}
		return o
	}

	// otherwise, return define the expected string values
	values := enum.GetValue()
	for _, v := range values {
		o.Enum = append(o.Enum, v.GetName())
	}
	o.Type = &openapi3.Types{openapi3.TypeString}

	return o
}

func (g *openapiGenerator) absoluteName(desc protomodel.CoreDesc) string {
	typeName := protomodel.DottedName(desc)
	return desc.PackageDesc().Name + "." + typeName
}

// converts the first section of the leading comment or the description of the proto
// to a single line of description.
func (g *openapiGenerator) generateDescription(desc protomodel.CoreDesc) string {
	if g.descriptionConfiguration.MultilineDescription {
		return g.generateMultiLineDescription(desc)
	}

	if !g.descriptionConfiguration.IncludeDescriptionInSchema {
		return ""
	}

	c := strings.TrimSpace(desc.Location().GetLeadingComments())
	t := strings.Split(c, "\n\n")[0]
	// omit the comment that starts with `$`.
	if strings.HasPrefix(t, "$") {
		return ""
	}

	return strings.Join(strings.Fields(t), " ")
}

func (g *openapiGenerator) generateMultiLineDescription(desc protomodel.CoreDesc) string {
	if !g.descriptionConfiguration.IncludeDescriptionInSchema {
		return ""
	}
	comments, _ := g.parseComments(desc)
	return comments
}

func (g *openapiGenerator) mustApplyRulesToSchema(
	rules []string,
	o *openapi3.Schema,
	target kubemarkers.TargetType,
) {
	if g.disableKubeMarkers {
		return
	}
	g.markerRegistry.MustApplyRulesToSchema(rules, o, target)
}

func (g *openapiGenerator) validationRules(desc protomodel.CoreDesc) []string {
	_, validationRules := g.parseComments(desc)
	return validationRules
}

func (g *openapiGenerator) parseComments(desc protomodel.CoreDesc) (comments string, validationRules []string) {
	c := strings.TrimSpace(desc.Location().GetLeadingComments())
	blocks := strings.Split(c, "\n\n")

	var ignoredKubeMarkersRegexp *regexp.Regexp
	if len(g.ignoredKubeMarkerSubstrings) > 0 {
		ignoredKubeMarkersRegexp = regexp.MustCompile(
			fmt.Sprintf("(?:%s)", strings.Join(g.ignoredKubeMarkerSubstrings, "|")),
		)
	}

	var sb strings.Builder
	for i, block := range blocks {
		if shouldNotRenderDesc(strings.TrimSpace(block)) {
			continue
		}
		if i > 0 {
			sb.WriteString("\n\n")
		}
		var blockSb strings.Builder
		lines := strings.Split(block, "\n")
		for i, line := range lines {
			if i > 0 {
				blockSb.WriteString("\n")
			}
			l := strings.TrimSpace(line)
			if shouldNotRenderDesc(l) {
				continue
			}

			if strings.HasPrefix(l, markers.Kubebuilder) {
				if isIgnoredKubeMarker(ignoredKubeMarkersRegexp, l) {
					continue
				}

				validationRules = append(validationRules, l)
				continue
			}
			if len(line) > 0 && line[0] == ' ' {
				line = line[1:]
			}
			blockSb.WriteString(strings.TrimRight(line, " "))
		}

		block = blockSb.String()
		sb.WriteString(block)
	}

	comments = strings.TrimSpace(sb.String())
	return
}

func shouldNotRenderDesc(desc string) bool {
	desc = strings.TrimSpace(desc)
	for _, marker := range descriptionExclusionMarkers {
		if strings.HasPrefix(desc, marker) {
			return true
		}
	}
	return false
}

func (g *openapiGenerator) fieldType(field *protomodel.FieldDescriptor) *openapi3.Schema {
	var schema *openapi3.Schema
	var isMap bool
	switch *field.Type {
	case descriptorpb.FieldDescriptorProto_TYPE_FLOAT, descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:
		schema = openapi3.NewFloat64Schema()

	case descriptorpb.FieldDescriptorProto_TYPE_INT32, descriptorpb.FieldDescriptorProto_TYPE_SINT32, descriptorpb.FieldDescriptorProto_TYPE_SFIXED32:
		schema = openapi3.NewInt32Schema()

	case descriptorpb.FieldDescriptorProto_TYPE_INT64, descriptorpb.FieldDescriptorProto_TYPE_SINT64,
		descriptorpb.FieldDescriptorProto_TYPE_SFIXED64, descriptorpb.FieldDescriptorProto_TYPE_FIXED64:
		if g.intNative {
			schema = openapi3.NewInt64Schema()
		} else {
			schema = g.generateSoloInt64Schema()
		}

	case descriptorpb.FieldDescriptorProto_TYPE_FIXED32:
		schema = openapi3.NewInt32Schema()

	case descriptorpb.FieldDescriptorProto_TYPE_UINT32:
		schema = openapi3.NewIntegerSchema().WithMin(0).WithMax(math.MaxUint32)

	case descriptorpb.FieldDescriptorProto_TYPE_UINT64:
		if g.intNative {
			// we don't set the max here beacause it is too large to be represented without scientific notation
			// in YAML format
			schema = openapi3.NewIntegerSchema().WithMin(0).WithFormat("uint64")
		} else {
			schema = g.generateSoloInt64Schema()
		}

	case descriptorpb.FieldDescriptorProto_TYPE_BOOL:
		schema = openapi3.NewBoolSchema()

	case descriptorpb.FieldDescriptorProto_TYPE_STRING:
		schema = openapi3.NewStringSchema()

	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		msg := field.FieldType.(*protomodel.MessageDescriptor)
		if soloSchema, ok := g.customSchemasByMessageName[g.absoluteName(msg)]; ok {
			// Allow for defining special Solo types
			schema = g.generateSoloMessageSchema(msg, &soloSchema)
		} else if msg.GetOptions().GetMapEntry() {
			isMap = true
			sr := g.fieldTypeRef(msg.Fields[1])
			if g.useRef && sr.Ref != "" {
				schema = openapi3.NewObjectSchema()
				// in `$ref`, the value of the schema is not in the output.
				sr.Value = nil
				schema.AdditionalProperties = openapi3.AdditionalProperties{Schema: sr}
			} else {
				schema = openapi3.NewObjectSchema().WithAdditionalProperties(sr.Value)
			}
		} else {
			schema = g.generateMessageSchema(msg)
		}

	case descriptorpb.FieldDescriptorProto_TYPE_BYTES:
		schema = openapi3.NewBytesSchema()

	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		enum := field.FieldType.(*protomodel.EnumDescriptor)
		schema = g.generateEnumSchema(enum)
	}

	if field.IsRepeated() && !isMap {
		schema = openapi3.NewArraySchema().WithItems(schema)
	}

	if schema != nil {
		schema.Description = g.generateDescription(field)
	}

	return schema
}

// fieldTypeRef generates the `$ref` in addition to the schema for a field.
func (g *openapiGenerator) fieldTypeRef(field *protomodel.FieldDescriptor) *openapi3.SchemaRef {
	s := g.fieldType(field)
	var ref string
	if *field.Type == descriptorpb.FieldDescriptorProto_TYPE_MESSAGE {
		msg := field.FieldType.(*protomodel.MessageDescriptor)
		// only generate `$ref` for top level messages.
		if _, ok := g.messages[g.relativeName(field.FieldType)]; ok && msg.Parent == nil {
			ref = fmt.Sprintf("#/components/schemas/%v", g.absoluteName(field.FieldType))
		}
	}
	return openapi3.NewSchemaRef(ref, s)
}

func (g *openapiGenerator) fieldName(field *protomodel.FieldDescriptor) string {
	return field.GetJsonName()
}

func (g *openapiGenerator) relativeName(desc protomodel.CoreDesc) string {
	typeName := protomodel.DottedName(desc)
	if desc.PackageDesc() == g.currentPackage {
		return typeName
	}

	return desc.PackageDesc().Name + "." + typeName
}

func isIgnoredKubeMarker(regexp *regexp.Regexp, l string) bool {
	if regexp == nil {
		return false
	}

	return regexp.MatchString(l)
}

