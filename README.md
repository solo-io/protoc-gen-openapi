
## What's this for?

`protoc-gen-openapi` is a plugin for the Google protocol buffer compiler to generate
openAPI V3 spec for any given input protobuf. It runs as a `protoc-gen-` binary that the
protobuf compiler infers from the `openapi_out` flag.

## Build `protoc-gen-openapi`

`protoc-gen-openapi` is written in Go, so ensure that is installed on your system. You
can follow the instructions on the [golang website](https://golang.org/doc/install) or
on Debian or Ubuntu, you can install it from the package manager:

```bash
sudo apt-get install -y golang
```

To build, run the following command from this project directory:

```bash
make build
```

Then ensure the resulting `protoc-gen-openapi` binary is in your `PATH`. A recommended location
is `$HOME/bin`:

```bash
cp _output/.bin/protoc-gen-openapi $HOME/bin
```

Since the following is often in your `$HOME/.bashrc` file:

```bash
export PATH=$HOME/bin:$PATH
```

## Using protoc-gen-openapi

---
**TIP**

The -I option in protoc is useful when you need to specify proto paths for imports.

---

Then to generate the OpenAPI spec of the protobuf defined by file.proto, run

```bash
protoc --openapi_out=output_directory input_directory/file.proto
```

With that input, the output will be written to

	output_directory/file.json

Other supported options are:
*   `per_file`
    *   when set to `true`, the output is per proto file instead of per package.
*   `single_file`
    *   when set to `true`, the output is a single file of all the input protos specified.
*   `use_ref`
    *   when set to `true`, the output uses the `$ref` field in OpenAPI spec to reference other schemas.
*   `yaml`
    *   when set to `true`, the output is in yaml file.
*   `include_description`
    *   when set to `true`, the openapi schema will include descriptions, generated from the proto message comment.
*   `multiline_description`
    *  when set to `true`, the openapi schema will include descriptions, generated from the proto message comment, that can span multiple lines. This can only be used with `yaml=true`.
*   `enum_as_int_or_string`
    *   when set to `true`, the openapi schema will include `x-kubernetes-int-or-string` on enums.
*   `additional_empty_schemas`
    *   a `+` separated list of message names (`core.solo.io.Status`), whose generated schema should be an empty object that accepts all values.
*  `proto_oneof`
    *   when set to `true`, the openapi schema will include `oneOf` emulating the behavior of proto `oneof`.
*  `int_native`
    *   when set to `true`, the native openapi schemas will be used for Integer types instead of Solo wrappers that add Kubernetes extension headers to the schema to treat int as strings.
*  `disable_kube_markers`
    *   when set to `true`, kubebuilder markers and validations such as PreserveUnknownFields, MinItems, default, and all CEL rules will be omitted from the OpenAPI schema. The Type and Required markers will be maintained.
*  `ignored_kube_markers`
    *   when set, ignores the contained kubebuilder markers and validations, and prevents them from being applied to the OpenAPI schema. 
