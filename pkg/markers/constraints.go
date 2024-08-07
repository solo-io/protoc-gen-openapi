package markers

import (
	"log"
	"math"

	"github.com/getkin/kin-openapi/openapi3"
)

// Maximum specifies the maximum numeric value that this field can have.
type Maximum float64

func (m Maximum) Value() float64 {
	return float64(m)
}

func (m Maximum) ApplyToSchema(o *openapi3.Schema) {
	if !hasNumericType(o) {
		log.Panicf("Maximum constraint applied to non-numeric type %s", o.Type)
	}
	o.WithMax(m.Value())
}

// Minimum specifies the minimum numeric value that this field can have. Negative numbers are supported.
type Minimum float64

func (m Minimum) Value() float64 {
	return float64(m)
}

func (m Minimum) ApplyToSchema(o *openapi3.Schema) {
	if !hasNumericType(o) {
		log.Panicf("must apply Minimum to a numeric type, got %s", o.Type)
	}
	o.WithMin(m.Value())
}

// ExclusiveMinimum indicates that the minimum is "up to" but not including that value.
type ExclusiveMinimum bool

func (m ExclusiveMinimum) ApplyToSchema(o *openapi3.Schema) {
	if !hasNumericType(o) {
		log.Panicf("must apply ExclusiveMinimum to a numeric type, got %s", o.Type)
	}
	o.WithExclusiveMin(bool(m))
}

// ExclusiveMaximum indicates that the maximum is "up to" but not including that value.
type ExclusiveMaximum bool

func (m ExclusiveMaximum) ApplyToSchema(o *openapi3.Schema) {
	if !hasNumericType(o) {
		log.Panicf("must apply ExclusiveMaximum to a numeric type, got %s", o.Type)
	}
	o.WithExclusiveMax(bool(m))
}

// MultipleOf specifies that this field must have a numeric value that's a multiple of this one.
type MultipleOf float64

func (m MultipleOf) Value() float64 {
	return float64(m)
}

func (m MultipleOf) ApplyToSchema(o *openapi3.Schema) {
	if !hasNumericType(o) {
		log.Panicf("must apply MultipleOf to a numeric type, got %s", o.Type)
	}
	if o.Type == "integer" && !isIntegral(m.Value()) {
		log.Panicf("cannot apply non-integral MultipleOf validation (%v) to integer value", m.Value())
	}
	val := m.Value()
	o.MultipleOf = &val
}

// MaxProperties restricts the number of keys in an object
type MaxProperties int

func (m MaxProperties) ApplyToSchema(o *openapi3.Schema) {
	if o.Type != "object" {
		log.Panicf("must apply MaxProperties to an object, got %s", o.Type)
	}
	o.WithMaxProperties(int64(m))
}

// MinProperties restricts the number of keys in an object
type MinProperties int

func (m MinProperties) ApplyToSchema(o *openapi3.Schema) {
	if o.Type != "object" {
		log.Panicf("must apply MinProperties to an object, got %s", o.Type)
	}
	o.WithMinProperties(int64(m))
}

// MaxLength specifies the maximum length for this string.
type MaxLength int

func (m MaxLength) ApplyToSchema(o *openapi3.Schema) {
	if o.Type != "string" {
		log.Panicf("must apply MaxLength to a string, got %s", o.Type)
	}
	o.WithMaxLength(int64(m))
}

// MinLength specifies the minimum length for this string.
type MinLength int

func (m MinLength) ApplyToSchema(o *openapi3.Schema) {
	if o.Type != "string" {
		log.Panicf("must apply MinLength to a string, got %s", o.Type)
	}
	o.WithMinLength(int64(m))
}

// Pattern specifies that this string must match the given regular expression.
type Pattern string

func (m Pattern) ApplyToSchema(o *openapi3.Schema) {
	if o.Type != "string" {
		log.Panicf("must apply Pattern to a string, got %s", o.Type)
	}
	o.WithPattern(string(m))
}

// MaxItems specifies the maximum length for this list.
type MaxItems int

func (m MaxItems) ApplyToSchema(o *openapi3.Schema) {
	if o.Type != "array" {
		log.Panicf("must apply MaxItems to an array, got %s", o.Type)
	}
	o.WithMaxItems(int64(m))
}

// MinItems specifies the minimum length for this list.
type MinItems int

func (m MinItems) ApplyToSchema(o *openapi3.Schema) {
	if o.Type != "array" {
		log.Panicf("must apply MinItems to an array, got %s", o.Type)
	}
	o.WithMinItems(int64(m))
}

// UniqueItems specifies that all items in this list must be unique.
type UniqueItems bool

func (m UniqueItems) ApplyToSchema(o *openapi3.Schema) {
	if o.Type != "array" {
		log.Panicf("must apply UniqueItems to an array, got %s", o.Type)
	}
	o.UniqueItems = bool(m)
}

// Enum specifies that this (scalar) field is restricted to the *exact* values specified here.
type Enum []interface{}

func (m Enum) ApplyToSchema(o *openapi3.Schema) {
	o.WithEnum(m...)
}

// Format specifies additional "complex" formatting for this field.
//
// For example, a date-time field would be marked as "type: string" and
// "format: date-time".
type Format string

func (m Format) ApplyToSchema(o *openapi3.Schema) {
	o.WithFormat(string(m))
}

// Type is a marker that specifies the type of a field in the schema
// It only supports the following types:
// - object: an opaque object
// - value: an opauqe value
type Type string

const (
	TypeObject Type = "object"
	TypeValue  Type = "value"
)

func (m Type) ApplyToSchema(o *openapi3.Schema) {
	// object and value types are special cased in the generator
	if o.Type == string(TypeObject) || o.Type != string(TypeValue) {
		return
	}
	o.Type = string(m)
}

// PreserveUnknownFields stops the apiserver from pruning fields which are not specified.
//
// By default the apiserver drops unknown fields from the request payload
// during the decoding step. This marker stops the API server from doing so.
// It affects fields recursively, but switches back to normal pruning behaviour
// if nested  properties or additionalProperties are specified in the schema.
// This can either be true or undefined. False
// is forbidden.
//
// Note: The kubebuilder:validation:XPreserveUnknownFields variant is deprecated
// in favor of the kubebuilder:pruning:PreserveUnknownFields variant.  They function
// identically.
type XPreserveUnknownFields struct{}

func (m XPreserveUnknownFields) ApplyToSchema(o *openapi3.Schema) {
	if o.ExtensionProps.Extensions == nil {
		o.ExtensionProps = openapi3.ExtensionProps{
			Extensions: map[string]interface{}{},
		}
	}
	o.ExtensionProps.Extensions["x-kubernetes-preserve-unknown-fields"] = true
}

// EmbeddedResource marks a fields as an embedded resource with apiVersion, kind and metadata fields.
//
// An embedded resource is a value that has apiVersion, kind and metadata fields.
// They are validated implicitly according to the semantics of the currently
// running apiserver. It is not necessary to add any additional schema for these
// field, yet it is possible. This can be combined with PreserveUnknownFields.
type XEmbeddedResource struct{}

func (m XEmbeddedResource) ApplyToSchema(o *openapi3.Schema) {
	if o.ExtensionProps.Extensions == nil {
		o.ExtensionProps = openapi3.ExtensionProps{
			Extensions: map[string]interface{}{},
		}
	}
	o.ExtensionProps.Extensions["x-kubernetes-embedded-resource"] = true
}

// IntOrString marks a fields as an IntOrString.
//
// This is required when applying patterns or other validations to an IntOrString
// field. Knwon information about the type is applied during the collapse phase
// and as such is not normally available during marker application.
type XIntOrString struct{}

func (m XIntOrString) ApplyToSchema(o *openapi3.Schema) {
	if o.ExtensionProps.Extensions == nil {
		o.ExtensionProps = openapi3.ExtensionProps{
			Extensions: map[string]interface{}{},
		}
	}
	o.ExtensionProps.Extensions["x-kubernetes-int-or-string"] = true
}

// XValidation marks a field as requiring a value for which a given
// expression evaluates to true.
//
// This marker may be repeated to specify multiple expressions, all of
// which must evaluate to true.
type XValidation struct {
	Rule              string `json:"rule"`
	Message           string `marker:",optional" json:"message,omitempty"`
	MessageExpression string `marker:",optional" json:"messageExpression,omitempty"`
}

func (x XValidation) ApplyToSchema(o *openapi3.Schema) {
	const validationsHeader = "x-kubernetes-validations"
	if o.ExtensionProps.Extensions == nil {
		o.ExtensionProps = openapi3.ExtensionProps{
			Extensions: map[string]interface{}{
				validationsHeader: []XValidation{},
			},
		}
	} else if o.ExtensionProps.Extensions[validationsHeader] == nil {
		o.ExtensionProps.Extensions[validationsHeader] = []XValidation{}
	}
	o.ExtensionProps.Extensions[validationsHeader] = append(o.ExtensionProps.Extensions[validationsHeader].([]XValidation), x)
}

// Nullable marks this field as allowing the "null" value.
//
// This is often not necessary, but may be helpful with custom serialization.
type Nullable struct{}

func (m Nullable) ApplyToSchema(o *openapi3.Schema) {
	o.WithNullable()
}

// Default sets the default value for this field.
//
// A default value will be accepted as any value valid for the
// field. Formatting for common types include: boolean: `true`, string:
// `Cluster`, numerical: `1.24`, array: `{1,2}`, object: `{policy:
// "delete"}`). Defaults should be defined in pruned form, and only best-effort
// validation will be performed. Full validation of a default requires
// submission of the containing CRD to an apiserver.
type Default struct {
	Value interface{}
}

func (m Default) ApplyToSchema(o *openapi3.Schema) {
	o.WithDefault(m.Value)
}

// Example sets the example value for this field.
//
// An example value will be accepted as any value valid for the
// field. Formatting for common types include: boolean: `true`, string:
// `Cluster`, numerical: `1.24`, array: `{1,2}`, object: `{policy:
// "delete"}`). Examples should be defined in pruned form, and only best-effort
// validation will be performed. Full validation of an example requires
// submission of the containing CRD to an apiserver.
type Example struct {
	Value interface{}
}

func (m Example) ApplyToSchema(o *openapi3.Schema) {
	o.Example = m.Value
}

type AltName string

func (a AltName) ApplyToSchema(o *openapi3.Schema) {}

// Schemaless marks a field as being a schemaless object.
//
// Schemaless objects are not introspected, so you must provide
// any type and validation information yourself. One use for this
// tag is for embedding fields that hold JSONSchema typed objects.
// Because this field disables all type checking, it is recommended
// to be used only as a last resort.
type Schemaless struct{}

func (m Schemaless) ApplyToSchema(o *openapi3.Schema) {
	// only preserve the description
	desc := o.Description
	nilSchema := openapi3.NewSchema()
	*o = *nilSchema
	o.Description = desc
}

// Required marks a field as required.
type Required struct{}

func (m Required) ApplyToSchema(o *openapi3.Schema) {
	// nothing to do, it is applied on the top level message containing the required field
}

func hasNumericType(o *openapi3.Schema) bool {
	return o.Type == "integer" || o.Type == "number"
}

func isIntegral(value float64) bool {
	return value == math.Trunc(value) && !math.IsNaN(value) && !math.IsInf(value, 0)
}
