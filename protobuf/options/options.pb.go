// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.8
// source: options.proto

package options

import (
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// FieldOptions expose custom configuration at the field level.
// The schema generator parses these options from the proto descriptors to generate appropriate OpenAPI schemas.
type FieldOptions struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Enable this option to treat this field as an unstructured object in the OpenAPI schema.
	// This can be set to disable infinite recursion of proto fields.
	TypeObject *bool `protobuf:"varint,1,opt,name=type_object,json=typeObject,proto3,oneof" json:"type_object,omitempty"`
	// Enable this option to treat this field as any value in the OpenAPI schema.
	// This can be set to disable infinite recursion when expanding references on recursive types.
	// In addition, should be set when using `google.protobuf.Value`, which can accept any JSON value type:
	// object, array, string, number, boolean, or null.
	TypeValue *bool `protobuf:"varint,2,opt,name=type_value,json=typeValue,proto3,oneof" json:"type_value,omitempty"`
}

func (x *FieldOptions) Reset() {
	*x = FieldOptions{}
	if protoimpl.UnsafeEnabled {
		mi := &file_options_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FieldOptions) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FieldOptions) ProtoMessage() {}

func (x *FieldOptions) ProtoReflect() protoreflect.Message {
	mi := &file_options_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FieldOptions.ProtoReflect.Descriptor instead.
func (*FieldOptions) Descriptor() ([]byte, []int) {
	return file_options_proto_rawDescGZIP(), []int{0}
}

func (x *FieldOptions) GetTypeObject() bool {
	if x != nil && x.TypeObject != nil {
		return *x.TypeObject
	}
	return false
}

func (x *FieldOptions) GetTypeValue() bool {
	if x != nil && x.TypeValue != nil {
		return *x.TypeValue
	}
	return false
}

var file_options_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptor.FieldOptions)(nil),
		ExtensionType: (*FieldOptions)(nil),
		Field:         123456,
		Name:          "solo.io.protocgenopenapi.options",
		Tag:           "bytes,123456,opt,name=options",
		Filename:      "options.proto",
	},
}

// Extension fields to descriptor.FieldOptions.
var (
	// TODO: update extension number once it has been registered with the upstream Extension registry
	//
	// optional solo.io.protocgenopenapi.FieldOptions options = 123456;
	E_Options = &file_options_proto_extTypes[0]
)

var File_options_proto protoreflect.FileDescriptor

var file_options_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x18, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x67,
	0x65, 0x6e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x1a, 0x20, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x65, 0x73, 0x63, 0x72,
	0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x77, 0x0a, 0x0c, 0x46,
	0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x24, 0x0a, 0x0b, 0x74,
	0x79, 0x70, 0x65, 0x5f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x48, 0x00, 0x52, 0x0a, 0x74, 0x79, 0x70, 0x65, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x88, 0x01,
	0x01, 0x12, 0x22, 0x0a, 0x0a, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x08, 0x48, 0x01, 0x52, 0x09, 0x74, 0x79, 0x70, 0x65, 0x56, 0x61, 0x6c,
	0x75, 0x65, 0x88, 0x01, 0x01, 0x42, 0x0e, 0x0a, 0x0c, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x6f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x5f, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x64, 0x0a, 0x07, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12,
	0x1d, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0xc0,
	0xc4, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x26, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x67, 0x65, 0x6e, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70,
	0x69, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x07,
	0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x88, 0x01, 0x01, 0x42, 0x38, 0x5a, 0x36, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e,
	0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_options_proto_rawDescOnce sync.Once
	file_options_proto_rawDescData = file_options_proto_rawDesc
)

func file_options_proto_rawDescGZIP() []byte {
	file_options_proto_rawDescOnce.Do(func() {
		file_options_proto_rawDescData = protoimpl.X.CompressGZIP(file_options_proto_rawDescData)
	})
	return file_options_proto_rawDescData
}

var file_options_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_options_proto_goTypes = []interface{}{
	(*FieldOptions)(nil),            // 0: solo.io.protocgenopenapi.FieldOptions
	(*descriptor.FieldOptions)(nil), // 1: google.protobuf.FieldOptions
}
var file_options_proto_depIdxs = []int32{
	1, // 0: solo.io.protocgenopenapi.options:extendee -> google.protobuf.FieldOptions
	0, // 1: solo.io.protocgenopenapi.options:type_name -> solo.io.protocgenopenapi.FieldOptions
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	1, // [1:2] is the sub-list for extension type_name
	0, // [0:1] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_options_proto_init() }
func file_options_proto_init() {
	if File_options_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_options_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FieldOptions); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_options_proto_msgTypes[0].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_options_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 1,
			NumServices:   0,
		},
		GoTypes:           file_options_proto_goTypes,
		DependencyIndexes: file_options_proto_depIdxs,
		MessageInfos:      file_options_proto_msgTypes,
		ExtensionInfos:    file_options_proto_extTypes,
	}.Build()
	File_options_proto = out.File
	file_options_proto_rawDesc = nil
	file_options_proto_goTypes = nil
	file_options_proto_depIdxs = nil
}