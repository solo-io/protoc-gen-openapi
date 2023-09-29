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

	"github.com/solo-io/protoc-gen-openapi/pkg/protocgen"
	"github.com/solo-io/protoc-gen-openapi/pkg/protomodel"

	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

func generate(request plugin.CodeGeneratorRequest) (*plugin.CodeGeneratorResponse, error) {
	options := newGenerationOptions()
	if err := options.parseParameters(request.GetParameter()); err != nil {
		return nil, err
	}

	m := protomodel.NewModel(&request, options.perFile)

	filesToGen := make(map[*protomodel.FileDescriptor]bool)
	for _, fileName := range request.FileToGenerate {
		fd := m.AllFilesByName[fileName]
		if fd == nil {
			return nil, fmt.Errorf("unable to find %s", request.FileToGenerate)
		}
		filesToGen[fd] = true
	}

	descriptionConfiguration := &DescriptionConfiguration{
		IncludeDescriptionInSchema: options.includeDescription,
	}

	g := newOpenAPIGenerator(options, m, descriptionConfiguration)
	return g.generateOutput(filesToGen)
}

func main() {
	protocgen.Generate(generate)
}
