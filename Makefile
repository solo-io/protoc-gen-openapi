
all: build run

build-and-test: build gotest

ROOTDIR := $(shell pwd)
OUTPUTDIR = $(ROOTDIR)/_output
BINDIR = $(OUTPUTDIR)/.bin

.PHONY: install-deps
install-deps: install-protoc
	mkdir -p $(BINDIR)
	GOBIN=$(BINDIR) go install github.com/golang/protobuf/protoc-gen-go

build-proto:
	mkdir -p $(ROOTDIR)/protobuf/options
	PATH=$(BINDIR):$(PATH) protoc -I=$(ROOTDIR)/protobuf/imports/ -I=$(ROOTDIR)/protobuf --go_out=$(OUTPUTDIR) $(ROOTDIR)/protobuf/options.proto
	cp $(OUTPUTDIR)/github.com/solo-io/protoc-gen-openapi/protobuf/options/options.pb.go $(ROOTDIR)/protobuf/options

build: install-deps build-proto
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/protoc-gen-openapi

run:
	rm -fr $(OUTPUTDIR)
	mkdir -p $(OUTPUTDIR)
	protoc --plugin=./$(BINDIR)/protoc-gen-openapi --openapi_out=single_file=true,use_ref=true:$(OUTPUTDIR)/. -Itestdata testdata/testpkg/test1.proto testdata/testpkg/test2.proto testdata/testpkg/test6.proto testdata/testpkg2/test3.proto

gotest:
	PATH=$(BINDIR):$(PATH) go test -v ./...

PROTOC_VERSION:=3.15.8
PROTOC_URL:=https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/protoc-${PROTOC_VERSION}
.PHONY: install-protoc
install-protoc:
	mkdir -p $(BINDIR)
	if [ $(shell ${BINDIR}/protoc --version | grep -c ${PROTOC_VERSION}) -ne 0 ]; then \
		echo expected protoc version ${PROTOC_VERSION} already installed ;\
	else \
		if [ "$(shell uname)" = "Darwin" ]; then \
			echo "downloading protoc for osx" ;\
			wget $(PROTOC_URL)-osx-x86_64.zip -O $(BINDIR)/protoc-${PROTOC_VERSION}.zip ;\
		elif [ "$(shell uname -m)" = "aarch64" ]; then \
			echo "downloading protoc for linux aarch64" ;\
			wget $(PROTOC_URL)-linux-aarch_64.zip -O $(BINDIR)/protoc-${PROTOC_VERSION}.zip ;\
		else \
			echo "downloading protoc for linux x86-64" ;\
			wget $(PROTOC_URL)-linux-x86_64.zip -O $(BINDIR)/protoc-${PROTOC_VERSION}.zip ;\
		fi ;\
		unzip $(BINDIR)/protoc-${PROTOC_VERSION}.zip -d $(BINDIR)/protoc-${PROTOC_VERSION} ;\
		mv $(BINDIR)/protoc-${PROTOC_VERSION}/bin/protoc $(BINDIR)/protoc ;\
		chmod +x $(BINDIR)/protoc ;\
		rm -rf $(BINDIR)/protoc-${PROTOC_VERSION} $(BINDIR)/protoc-${PROTOC_VERSION}.zip ;\
	fi

clean:
	@rm -fr $(OUTPUTDIR) $(BINDIR)/protoc-gen-openapi
