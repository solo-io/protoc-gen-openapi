package markers

import (
	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

const (
	validationsHeader = "x-kubernetes-validations"
)

var _ SchemaMarker = XValidation{}

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

func (XValidation) Help() *markers.DefinitionHelp {
	return &markers.DefinitionHelp{
		Category: "CRD validation",
		DetailedHelp: markers.DetailedHelp{
			Summary: "marks a field as requiring a value for which a given expression evaluates to true. ",
			Details: "This marker may be repeated to specify multiple expressions, all of which must evaluate to true.",
		},
		FieldHelp: map[string]markers.DetailedHelp{
			"Rule": {
				Summary: "",
				Details: "",
			},
			"Message": {
				Summary: "",
				Details: "",
			},
		},
	}
}
