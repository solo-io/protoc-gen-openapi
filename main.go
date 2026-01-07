// Copyright 2019 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
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
	"fmt"
	"strings"

	"github.com/solo-io/protoc-gen-openapi/pkg/protocgen"
	"github.com/solo-io/protoc-gen-openapi/pkg/protomodel"

	"google.golang.org/protobuf/types/pluginpb"
)

// Breaks the comma-separated list of key=value pairs
// in the parameter string into an easy to use map.
func extractParams(parameter string) map[string]string {
	m := make(map[string]string)
	for _, p := range strings.Split(parameter, ",") {
		if p == "" {
			continue
		}

		if i := strings.Index(p, "="); i < 0 {
			m[p] = ""
		} else {
			m[p[0:i]] = p[i+1:]
		}
	}

	return m
}

func generate(request pluginpb.CodeGeneratorRequest) (*pluginpb.CodeGeneratorResponse, error) {
	perFile := false
	singleFile := false
	yaml := false
	useRef := false
	includeDescription := true
	multilineDescription := false
	enumAsIntOrString := false
	enumAsInt := false
	protoOneof := false
	intNative := false
	disableKubeMarkers := false

	var enumNamesExtensions []string
	var messagesWithEmptySchema []string
	var ignoredKubeMarkerSubstrings []string

	p := extractParams(request.GetParameter())
	for k, v := range p {
		if k == "per_file" {
			switch strings.ToLower(v) {
			case "true":
				perFile = true
			case "false":
				perFile = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for per_file", v)
			}
		} else if k == "single_file" {
			switch strings.ToLower(v) {
			case "true":
				if perFile {
					return nil, fmt.Errorf("output is already to be generated per file, cannot output to a single file")
				}
				singleFile = true
			case "false":
				singleFile = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for single_file", v)
			}
		} else if k == "yaml" {
			yaml = true
		} else if k == "use_ref" {
			switch strings.ToLower(v) {
			case "true":
				useRef = true
			case "false":
				useRef = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for use_ref", v)
			}
		} else if k == "include_description" {
			switch strings.ToLower(v) {
			case "true":
				includeDescription = true
			case "false":
				includeDescription = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for include_description", v)
			}
		} else if k == "multiline_description" {
			switch strings.ToLower(v) {
			case "true":
				multilineDescription = true
			case "false":
				multilineDescription = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for multiline_description", v)
			}
		} else if k == "enum_as_int_or_string" {
			switch strings.ToLower(v) {
			case "true":
				enumAsIntOrString = true
			case "false":
				enumAsIntOrString = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for enum_as_int_or_string", v)
			}
		} else if k == "use_int_enums" {
			switch strings.ToLower(v) {
			case "true":
				enumAsInt = true
			case "false":
				enumAsInt = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for enum_as_int", v)
			}
		} else if k == "enum_names_extensions" {
			enumNamesExtensions = strings.Split(v, ";")
			if len(enumNamesExtensions) == 0 {
				return nil, fmt.Errorf("cant use '%s' as enum_names_extensions, provide a semicolon separated list", v)
			}
		} else if k == "proto_oneof" {
			switch strings.ToLower(v) {
			case "true":
				protoOneof = true
			case "false":
				protoOneof = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for proto_oneof", v)
			}
		} else if k == "int_native" {
			switch strings.ToLower(v) {
			case "true":
				intNative = true
			case "false":
				intNative = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for int_native", v)
			}
		} else if k == "additional_empty_schema" {
			messagesWithEmptySchema = strings.Split(v, "+")
		} else if k == "disable_kube_markers" {
			switch strings.ToLower(v) {
			case "true":
				disableKubeMarkers = true
			case "false":
				disableKubeMarkers = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for disable_kube_markers", v)
			}
		} else if k == "ignored_kube_marker_substrings" {
			if len(v) > 0 {
				ignoredKubeMarkerSubstrings = strings.Split(v, "+")
			}
		} else {
			return nil, fmt.Errorf("unknown argument '%s' specified", k)
		}
	}

	if !yaml && multilineDescription {
		return nil, fmt.Errorf("multiline_description is only supported when yaml=true")
	}

	m := protomodel.NewModel(&request, perFile)

	filesToGen := make(map[*protomodel.FileDescriptor]bool)
	for _, fileName := range request.FileToGenerate {
		fd := m.AllFilesByName[fileName]
		if fd == nil {
			return nil, fmt.Errorf("unable to find %s", request.FileToGenerate)
		}
		filesToGen[fd] = true
	}

	descriptionConfiguration := &DescriptionConfiguration{
		IncludeDescriptionInSchema: includeDescription,
		MultilineDescription:       multilineDescription,
	}

	g := newOpenAPIGenerator(
		m,
		perFile,
		singleFile,
		yaml,
		useRef,
		descriptionConfiguration,
		enumAsIntOrString,
		enumAsInt,
		enumNamesExtensions,
		messagesWithEmptySchema,
		protoOneof,
		intNative,
		disableKubeMarkers,
		ignoredKubeMarkerSubstrings,
	)
	return g.generateOutput(filesToGen)
}

func main() {
	protocgen.Generate(generate)
}
