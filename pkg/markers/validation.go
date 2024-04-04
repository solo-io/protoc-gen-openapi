package markers

import (
	"fmt"
	"log"

	"github.com/getkin/kin-openapi/openapi3"
	"sigs.k8s.io/controller-tools/pkg/markers"
)

var ValidationMarkers = mustMakeAllWithPrefix("kubebuilder:validation", markers.DescribesField,
	XValidation{},
	Type(""),
)

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

	// AllDefinitions = append(AllDefinitions, FieldOnlyMarkers...)
	// AllDefinitions = append(AllDefinitions, ValidationIshMarkers...)
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
		defn := r.mRegistry.Lookup(rule, TargetType)
		if defn == nil {
			return fmt.Errorf("no definition found for rule: %s", rule)
		}
		val, err := defn.Parse(rule)
		if err != nil {
			return fmt.Errorf("error parsing rule: %s", err)
		}
		if s, ok := val.(SchemaMarker); ok {
			s.ApplyToSchema(o)
		} else {
			return fmt.Errorf("expected SchemaMarker, got %T", val)
		}
	}
	return nil
}

func (r *Registry) GetSchemaType(
	rules []string,
) Type {
	for _, rule := range rules {
		defn := r.mRegistry.Lookup(rule, TargetType)
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
