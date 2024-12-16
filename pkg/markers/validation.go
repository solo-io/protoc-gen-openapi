package markers

import (
	"fmt"
	"log"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

const (
	// Kubebuilder marker used in comments
	Kubebuilder = "+kubebuilder:"

	FieldRequired = "required"
	FieldOptional = "optional"
)

var (
	_ SchemaMarker = Maximum(0)
	_ SchemaMarker = Minimum(0)
	_ SchemaMarker = ExclusiveMaximum(false)
	_ SchemaMarker = ExclusiveMinimum(false)
	_ SchemaMarker = MultipleOf(0)
	_ SchemaMarker = MinProperties(0)
	_ SchemaMarker = MaxProperties(0)
	_ SchemaMarker = MaxLength(0)
	_ SchemaMarker = MinLength(0)
	_ SchemaMarker = Pattern("")
	_ SchemaMarker = MaxItems(0)
	_ SchemaMarker = MinItems(0)
	_ SchemaMarker = UniqueItems(false)
	_ SchemaMarker = Enum(nil)
	_ SchemaMarker = Format("")
	_ SchemaMarker = Type("")
	_ SchemaMarker = XPreserveUnknownFields{}
	_ SchemaMarker = XEmbeddedResource{}
	_ SchemaMarker = XIntOrString{}
	_ SchemaMarker = XValidation{}
)

// ValidationMarkers lists all available markers that affect CRD schema generation,
// except for the few that don't make sense as type-level markers (see FieldOnlyMarkers).
// All markers start with `+kubebuilder:validation:`, and continue with their type name.
// A copy is produced of all markers that describes types as well, for making types
// reusable and writing complex validations on slice items.
var ValidationMarkers = mustMakeAllWithPrefix("kubebuilder:validation", markers.DescribesField,
	// Numeric markers
	Maximum(0),
	Minimum(0),
	ExclusiveMaximum(false),
	ExclusiveMinimum(false),
	MultipleOf(0),
	MinProperties(0),
	MaxProperties(0),

	// string markers
	MaxLength(0),
	MinLength(0),
	Pattern(""),

	// Slice markers
	MaxItems(0),
	MinItems(0),
	UniqueItems(false),

	// General markers
	Enum(nil),
	Format(""),
	Type(""),
	XPreserveUnknownFields{},
	XEmbeddedResource{},
	XIntOrString{},
	XValidation{},
)

// FieldOnlyMarkers list field-specific validation markers (i.e. those markers that don't make
// sense on a type, and thus aren't in ValidationMarkers).
var FieldOnlyMarkers = []*definitionWithHelp{
	must(markers.MakeDefinition("kubebuilder:validation:Required", markers.DescribesField, Required{})).
		WithHelp(markers.SimpleHelp("CRD validation", "specifies that this field is required, if fields are optional by default.")),

	// must(markers.MakeDefinition("kubebuilder:validation:Optional", markers.DescribesField, struct{}{})).
	// 	WithHelp(markers.SimpleHelp("CRD validation", "specifies that this field is optional, if fields are required by default.")),

	must(markers.MakeDefinition("kubebuilder:validation:Nullable", markers.DescribesField, Nullable{})),

	must(markers.MakeAnyTypeDefinition("kubebuilder:default", markers.DescribesField, Default{})),

	must(markers.MakeAnyTypeDefinition("kubebuilder:example", markers.DescribesField, Example{})),

	must(markers.MakeDefinition("kubebuilder:validation:EmbeddedResource", markers.DescribesField, XEmbeddedResource{})),

	must(markers.MakeDefinition("kubebuilder:validation:Schemaless", markers.DescribesField, Schemaless{})),
}

// ValidationIshMarkers are field-and-type markers that don't fall under the
// :validation: prefix, and/or don't have a name that directly matches their
// type.
var ValidationIshMarkers = []*definitionWithHelp{
	must(markers.MakeDefinition("kubebuilder:pruning:PreserveUnknownFields", markers.DescribesField, XPreserveUnknownFields{})),
	must(markers.MakeDefinition("kubebuilder:pruning:PreserveUnknownFields", markers.DescribesType, XPreserveUnknownFields{})),
}

type SchemaMarker interface {
	ApplyToSchema(o *openapi3.Schema)
}

func init() {
	AllDefinitions = append(AllDefinitions, ValidationMarkers...)

	for _, def := range ValidationMarkers {
		newDef := *def.Definition
		// copy both parts so we don't change the definition
		typDef := definitionWithHelp{
			Definition: &newDef,
			Help:       def.Help,
		}
		typDef.Target = markers.DescribesType
		AllDefinitions = append(AllDefinitions, &typDef)
	}

	AllDefinitions = append(AllDefinitions, FieldOnlyMarkers...)
	AllDefinitions = append(AllDefinitions, ValidationIshMarkers...)
}

func (r *Registry) MustApplyRulesToSchema(
	rules []string,
	o *openapi3.Schema,
	target markers.TargetType,
) {
	err := r.ApplyRulesToSchema(rules, o, target)
	if err != nil {
		log.Panicf("error applying rules to schema: %s", err)
	}
}

func (r *Registry) ApplyRulesToSchema(
	rules []string,
	o *openapi3.Schema,
	target markers.TargetType,
) error {
	for _, rule := range rules {
		var (
			transformedRule string
			targetSchema    *openapi3.Schema
			err             error
		)

		if isItemRule(rule) {
			transformedRule, targetSchema, err = r.transformItemRule(rule, o)
		} else {
			// Regular (non-items) rule
			transformedRule = rule
			targetSchema = o
		}

		if err != nil {
			return err
		}

		defn := r.mRegistry.Lookup(transformedRule, target)
		if defn == nil {
			return fmt.Errorf("no definition found for rule: %s", transformedRule)
		}

		val, err := defn.Parse(transformedRule)
		if err != nil {
			return fmt.Errorf("error parsing rule: %s", err)
		}

		if s, ok := val.(SchemaMarker); ok {
			s.ApplyToSchema(targetSchema)
		} else {
			return fmt.Errorf("expected SchemaMarker, got %T for rule %s", val, transformedRule)
		}
	}

	return nil
}

// isItemRule checks if the given rule applies to items.
func isItemRule(rule string) bool {
	return strings.Contains(rule, ":items:")
}

// transformItemRule transforms an item rule into a standard-looking rule
// and returns the appropriate target schema (the items schema).
func (r *Registry) transformItemRule(rule string, o *openapi3.Schema) (string, *openapi3.Schema, error) {
	parts := strings.SplitN(rule, ":items:", 2)
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid item rule format: %s", rule)
	}

	// Transform the item rule into a standard rule
	// e.g. "+kubebuilder:validation:items:MaxLength=255" -> "+kubebuilder:validation:MaxLength=255"
	transformedRule := parts[0] + ":" + parts[1]

	if o.Type != "array" {
		return "", nil, fmt.Errorf("items validation rule applied to non-array type: %s", o.Type)
	}
	if o.Items == nil {
		o.Items = openapi3.NewSchemaRef("", &openapi3.Schema{})
	}

	return transformedRule, o.Items.Value, nil
}

func (r *Registry) GetSchemaType(
	rules []string,
	target markers.TargetType,
) Type {
	for _, rule := range rules {
		defn := r.mRegistry.Lookup(rule, target)
		if defn == nil {
			log.Panicf("no definition found for rule: %s", rule)
		}
		val, err := defn.Parse(rule)
		if err != nil {
			log.Panicf("error parsing rule: %s", err)
		}
		if s, ok := val.(Type); ok {
			return s
		}
	}
	return ""
}

func (r *Registry) IsRequired(
	rules []string,
) bool {
	for _, rule := range rules {
		defn := r.mRegistry.Lookup(rule, markers.DescribesField)
		if defn == nil {
			log.Panicf("no definition found for rule: %s", rule)
		}
		if strings.HasPrefix(rule, "+kubebuilder:validation:Required") {
			return true
		}
	}
	return false
}
