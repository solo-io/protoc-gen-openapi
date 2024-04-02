package markers

import (
	"fmt"
	"log"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	validationKey     = "+kubebuilder:validation:"
	validationsHeader = "x-kubernetes-validations"

	ruleDelimiter = ";;"
)

const (
	TypeObject = "object"
	TypeValue  = "value"
)

var (
	typeKey        = validationKey + "Type="
	xValidationKey = validationKey + "XValidation:"
)

type ValidationRule struct {
	Rule                     string `json:"rule"`
	Message                  string `json:"message,omitempty"`
	MessageMessageExpression string `json:"messageExpression,omitempty"`
}

func ParseType(rules []string) string {
	for _, rule := range rules {
		if strings.HasPrefix(rule, typeKey) {
			return strings.TrimPrefix(rule, typeKey)
		}
	}
	return ""
}

func ApplyToSchema(o *openapi3.Schema, rules []string) {
	for _, rule := range rules {
		err := applyRule(o, rule)
		if err != nil {
			log.Panicf("error applying rule: %v", err)
		}
	}
}

func applyRule(o *openapi3.Schema, rule string) error {
	rule = strings.TrimSpace(rule)

	switch {
	case strings.HasPrefix(rule, xValidationKey):
		rule, err := parseXValidationRule(strings.TrimPrefix(rule, xValidationKey))
		if err != nil {
			return err
		}
		applyXValidationRule(o, rule)

	case strings.HasPrefix(rule, typeKey):
		// ignore, already handled in ParseType

	default:
		return fmt.Errorf("unsupported validation rule: %s", rule)
	}

	return nil
}

func parseXValidationRule(ruleStr string) (ValidationRule, error) {
	parts := strings.Split(ruleStr, ruleDelimiter)
	var rule ValidationRule
	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasPrefix(part, "rule="):
			rule.Rule = extractQuoted(strings.TrimPrefix(part, "rule="))
		case strings.HasPrefix(part, "message="):
			rule.Message = extractQuoted(strings.TrimPrefix(part, "message="))
		case strings.HasPrefix(part, "messageExpression="):
			rule.MessageMessageExpression = extractQuoted(strings.TrimPrefix(part, "messageExpression="))
		}
	}
	if rule.Rule == "" {
		return rule, fmt.Errorf("missing 'rule' field in rule: %s", ruleStr)
	}
	return rule, nil
}

func extractQuoted(s string) string {
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	log.Panicf("validation rule fields must be quoted; got: %s", s)
	return ""
}

func applyXValidationRule(o *openapi3.Schema, rule ValidationRule) {
	if o.ExtensionProps.Extensions == nil {
		o.ExtensionProps = openapi3.ExtensionProps{
			Extensions: map[string]interface{}{
				validationsHeader: []ValidationRule{},
			},
		}
	} else if o.ExtensionProps.Extensions[validationsHeader] == nil {
		o.ExtensionProps.Extensions[validationsHeader] = []ValidationRule{}
	}
	o.ExtensionProps.Extensions[validationsHeader] = append(o.ExtensionProps.Extensions[validationsHeader].([]ValidationRule), rule)
}
