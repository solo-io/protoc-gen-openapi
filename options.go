package main

import (
	"fmt"
	"strconv"
	"strings"
)

type generationOptions struct {
	perFile                 bool
	singleFile              bool
	yaml                    bool
	useRef                  bool
	includeDescription      bool
	enumAsIntOrString       bool
	messagesWithEmptySchema []string
	strictProto3Optional    bool
}

func newGenerationOptions() *generationOptions {
	return &generationOptions{
		includeDescription: true,
	}
}

func (o *generationOptions) parseParameters(args string) error {
	p := extractParams(args)
	for k, v := range p {
		if k == "per_file" {
			if val, err := strconv.ParseBool(v); err != nil {
				return fmt.Errorf("unknown value '%s' for per_file", v)
			} else {
				o.perFile = val
			}
		} else if k == "single_file" {
			if val, err := strconv.ParseBool(v); err != nil {
				return fmt.Errorf("unknown value '%s' for single_file", v)
			} else {
				o.singleFile = val
			}
			if o.perFile {
				return fmt.Errorf("output is already to be generated per file, cannot output to a single file")
			}
		} else if k == "yaml" {
			o.yaml = true
		} else if k == "use_ref" {
			if val, err := strconv.ParseBool(v); err != nil {
				return fmt.Errorf("unknown value '%s' for use_ref", v)
			} else {
				o.useRef = val
			}
		} else if k == "include_description" {
			if val, err := strconv.ParseBool(v); err != nil {
				return fmt.Errorf("unknown value '%s' for include_description", v)
			} else {
				o.includeDescription = val
			}
		} else if k == "enum_as_int_or_string" {
			if val, err := strconv.ParseBool(v); err != nil {
				return fmt.Errorf("unknown value '%s' for enum_as_int_or_string", v)
			} else {
				o.enumAsIntOrString = val
			}
		} else if k == "additional_empty_schema" {
			o.messagesWithEmptySchema = strings.Split(v, "+")
		} else if k == "strict_proto3_optional" {
			if val, err := strconv.ParseBool(v); err != nil {
				return fmt.Errorf("unknown value '%s' for strict_proto3_optional", v)
			} else {
				o.strictProto3Optional = val
			}
		} else {
			return fmt.Errorf("unknown argument '%s' specified", k)
		}
	}

	return nil
}

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
