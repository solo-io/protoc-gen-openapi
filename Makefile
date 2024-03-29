
all: build run

ROOTDIR := $(shell pwd)
OUTPUTDIR = $(ROOTDIR)/_output
DEPSGOBIN = $(OUTPUTDIR)/.bin

.PHONY: install-deps
install-deps:
	mkdir -p $(DEPSGOBIN)
	GOBIN=$(DEPSGOBIN) go install github.com/golang/protobuf/protoc-gen-go

build-proto:
	mkdir -p $(ROOTDIR)/protobuf/options
	PATH=$(DEPSGOBIN):$(PATH) protoc -I=$(ROOTDIR)/protobuf/imports/ -I=$(ROOTDIR)/protobuf --go_out=$(OUTPUTDIR) $(ROOTDIR)/protobuf/options.proto
	cp $(OUTPUTDIR)/github.com/solo-io/protoc-gen-openapi/protobuf/options/options.pb.go $(ROOTDIR)/protobuf/options

build: install-deps build-proto
	mkdir -p $(DEPSGOBIN)
	go build -o $(DEPSGOBIN)/protoc-gen-openapi

run:
	rm -fr $(OUTPUTDIR)
	mkdir -p $(OUTPUTDIR)
	protoc --plugin=./$(DEPSGOBIN)/protoc-gen-openapi --openapi_out=single_file=true,use_ref=true:$(OUTPUTDIR)/. -Itestdata testdata/testpkg/test1.proto testdata/testpkg/test2.proto testdata/testpkg/test6.proto testdata/testpkg2/test3.proto

gotest:
	go test

clean:
	@rm -fr $(OUTPUTDIR) $(DEPSGOBIN)/protoc-gen-openapi
