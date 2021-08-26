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
	"strconv"
	"strings"

	"github.com/sam-heilbron/protoc-gen-openapi/pkg/protocgen"
	"github.com/sam-heilbron/protoc-gen-openapi/pkg/protomodel"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
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

func generate(request plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	perFile := false
	singleFile := false
	yaml := false
	useRef := false
	maxCharactersInDescription := 0
	includeDescription := true
	enumAsIntOrString := false

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
		} else if k == "max_description_characters" {
			var err error
			maxCharactersInDescription, err = strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("unknown value '%s' for max_description_characters", v)
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
		} else if k == "enum_as_int_or_string" {
			switch strings.ToLower(v) {
			case "true":
				enumAsIntOrString = true
			case "false":
				enumAsIntOrString = false
			default:
				return nil, fmt.Errorf("unknown value '%s' for enum_as_int_or_string", v)
			}
		} else {
			return nil, fmt.Errorf("unknown argument '%s' specified", k)
		}
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
		MaxDescriptionCharacters:   maxCharactersInDescription,
		IncludeDescriptionInSchema: includeDescription,
	}

	g := newOpenAPIGenerator(m, perFile, singleFile, yaml, useRef, descriptionConfiguration, enumAsIntOrString)
	return g.generateOutput(filesToGen)
}

func main() {
	protocgen.Generate(generate)
}
