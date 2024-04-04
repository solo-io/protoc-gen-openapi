//go:build tools
// +build tools

// Explanation for tools pattern:
// https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

package tools

import (
	_ "github.com/golang/protobuf/protoc-gen-go"
)
