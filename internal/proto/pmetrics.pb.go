// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v3.15.8
// source: internal/proto/pmetrics.proto

package proto

import (
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

type MetricRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metric []byte `protobuf:"bytes,1,opt,name=metric,proto3" json:"metric,omitempty"` // метрика в JSON формате
}

func (x *MetricRequest) Reset() {
	*x = MetricRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pmetrics_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricRequest) ProtoMessage() {}

func (x *MetricRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pmetrics_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricRequest.ProtoReflect.Descriptor instead.
func (*MetricRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_pmetrics_proto_rawDescGZIP(), []int{0}
}

func (x *MetricRequest) GetMetric() []byte {
	if x != nil {
		return x.Metric
	}
	return nil
}

type MetricResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Error string `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *MetricResponse) Reset() {
	*x = MetricResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pmetrics_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MetricResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MetricResponse) ProtoMessage() {}

func (x *MetricResponse) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pmetrics_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MetricResponse.ProtoReflect.Descriptor instead.
func (*MetricResponse) Descriptor() ([]byte, []int) {
	return file_internal_proto_pmetrics_proto_rawDescGZIP(), []int{1}
}

func (x *MetricResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type EncMetric struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data0 string `protobuf:"bytes,1,opt,name=data0,proto3" json:"data0,omitempty"` // зашифрованный сеансовый ключ
	Data  string `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`   // зашифрованные данные
}

func (x *EncMetric) Reset() {
	*x = EncMetric{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pmetrics_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EncMetric) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EncMetric) ProtoMessage() {}

func (x *EncMetric) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pmetrics_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EncMetric.ProtoReflect.Descriptor instead.
func (*EncMetric) Descriptor() ([]byte, []int) {
	return file_internal_proto_pmetrics_proto_rawDescGZIP(), []int{2}
}

func (x *EncMetric) GetData0() string {
	if x != nil {
		return x.Data0
	}
	return ""
}

func (x *EncMetric) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

type EncMetricRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data *EncMetric `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *EncMetricRequest) Reset() {
	*x = EncMetricRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pmetrics_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *EncMetricRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EncMetricRequest) ProtoMessage() {}

func (x *EncMetricRequest) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pmetrics_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EncMetricRequest.ProtoReflect.Descriptor instead.
func (*EncMetricRequest) Descriptor() ([]byte, []int) {
	return file_internal_proto_pmetrics_proto_rawDescGZIP(), []int{3}
}

func (x *EncMetricRequest) GetData() *EncMetric {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_internal_proto_pmetrics_proto protoreflect.FileDescriptor

var file_internal_proto_pmetrics_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x70, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x08, 0x70, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x22, 0x27, 0x0a, 0x0d, 0x4d, 0x65, 0x74,
	0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x6d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x22, 0x26, 0x0a, 0x0e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x35, 0x0a, 0x09, 0x45, 0x6e,
	0x63, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x14, 0x0a, 0x05, 0x64, 0x61, 0x74, 0x61, 0x30,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x64, 0x61, 0x74, 0x61, 0x30, 0x12, 0x12, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x22, 0x3b, 0x0a, 0x10, 0x45, 0x6e, 0x63, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x27, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x70, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x45,
	0x6e, 0x63, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x32, 0xa7,
	0x02, 0x0a, 0x07, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x41, 0x0a, 0x0c, 0x52, 0x65,
	0x70, 0x6f, 0x72, 0x74, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x12, 0x17, 0x2e, 0x70, 0x6d, 0x65,
	0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x70, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x47, 0x0a,
	0x0f, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x45, 0x6e, 0x63, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63,
	0x12, 0x1a, 0x2e, 0x70, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x45, 0x6e, 0x63, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x70,
	0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x44, 0x0a, 0x0d, 0x52, 0x65, 0x70, 0x6f, 0x72, 0x74,
	0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x17, 0x2e, 0x70, 0x6d, 0x65, 0x74, 0x72, 0x69,
	0x63, 0x73, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x18, 0x2e, 0x70, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x4d, 0x65, 0x74, 0x72,
	0x69, 0x63, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x12, 0x4a, 0x0a, 0x10,
	0x52, 0x65, 0x70, 0x6f, 0x72, 0x74, 0x45, 0x6e, 0x63, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73,
	0x12, 0x1a, 0x2e, 0x70, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x45, 0x6e, 0x63, 0x4d,
	0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x70,
	0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x28, 0x01, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x68, 0x72, 0x61, 0x70, 0x6f, 0x76, 0x64, 0x31, 0x2f,
	0x70, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61,
	0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_proto_pmetrics_proto_rawDescOnce sync.Once
	file_internal_proto_pmetrics_proto_rawDescData = file_internal_proto_pmetrics_proto_rawDesc
)

func file_internal_proto_pmetrics_proto_rawDescGZIP() []byte {
	file_internal_proto_pmetrics_proto_rawDescOnce.Do(func() {
		file_internal_proto_pmetrics_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_proto_pmetrics_proto_rawDescData)
	})
	return file_internal_proto_pmetrics_proto_rawDescData
}

var file_internal_proto_pmetrics_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_internal_proto_pmetrics_proto_goTypes = []interface{}{
	(*MetricRequest)(nil),    // 0: pmetrics.MetricRequest
	(*MetricResponse)(nil),   // 1: pmetrics.MetricResponse
	(*EncMetric)(nil),        // 2: pmetrics.EncMetric
	(*EncMetricRequest)(nil), // 3: pmetrics.EncMetricRequest
}
var file_internal_proto_pmetrics_proto_depIdxs = []int32{
	2, // 0: pmetrics.EncMetricRequest.data:type_name -> pmetrics.EncMetric
	0, // 1: pmetrics.Metrics.ReportMetric:input_type -> pmetrics.MetricRequest
	3, // 2: pmetrics.Metrics.ReportEncMetric:input_type -> pmetrics.EncMetricRequest
	0, // 3: pmetrics.Metrics.ReportMetrics:input_type -> pmetrics.MetricRequest
	3, // 4: pmetrics.Metrics.ReportEncMetrics:input_type -> pmetrics.EncMetricRequest
	1, // 5: pmetrics.Metrics.ReportMetric:output_type -> pmetrics.MetricResponse
	1, // 6: pmetrics.Metrics.ReportEncMetric:output_type -> pmetrics.MetricResponse
	1, // 7: pmetrics.Metrics.ReportMetrics:output_type -> pmetrics.MetricResponse
	1, // 8: pmetrics.Metrics.ReportEncMetrics:output_type -> pmetrics.MetricResponse
	5, // [5:9] is the sub-list for method output_type
	1, // [1:5] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_internal_proto_pmetrics_proto_init() }
func file_internal_proto_pmetrics_proto_init() {
	if File_internal_proto_pmetrics_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_proto_pmetrics_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetricRequest); i {
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
		file_internal_proto_pmetrics_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MetricResponse); i {
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
		file_internal_proto_pmetrics_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EncMetric); i {
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
		file_internal_proto_pmetrics_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*EncMetricRequest); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_proto_pmetrics_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_internal_proto_pmetrics_proto_goTypes,
		DependencyIndexes: file_internal_proto_pmetrics_proto_depIdxs,
		MessageInfos:      file_internal_proto_pmetrics_proto_msgTypes,
	}.Build()
	File_internal_proto_pmetrics_proto = out.File
	file_internal_proto_pmetrics_proto_rawDesc = nil
	file_internal_proto_pmetrics_proto_goTypes = nil
	file_internal_proto_pmetrics_proto_depIdxs = nil
}
