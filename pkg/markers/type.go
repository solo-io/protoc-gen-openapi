package markers

import (
	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

// Type is a marker that specifies the type of a field in the schema
// It only supports the following types:
// - object: an opaque object
// - value: an opauqe value
type Type string

const (
	TypeObject Type = "object"
	TypeValue  Type = "value"
)

func (Type) ApplyToSchema(o *openapi3.Schema) {
	// we don't need Type mutations on the schema yet
}

func (Type) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "CRD validation",
		DetailedHelp: markers.DetailedHelp{
			Summary: "overrides the type for this field (which defaults to the equivalent of the proto type). ",
			Details: "This is only used to .",
		},
		FieldHelp: map[string]markers.DetailedHelp{},
	}
}
