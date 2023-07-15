// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v3.21.12
// source: zns.proto

package protos

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

type Segment struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SegmentNo   int32  `protobuf:"varint,1,opt,name=segment_no,json=segmentNo,proto3" json:"segment_no,omitempty"`
	Map         []byte `protobuf:"bytes,2,opt,name=map,proto3" json:"map,omitempty"`
	SegmentType int32  `protobuf:"varint,3,opt,name=segment_type,json=segmentType,proto3" json:"segment_type,omitempty"`
}

func (x *Segment) Reset() {
	*x = Segment{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zns_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Segment) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Segment) ProtoMessage() {}

func (x *Segment) ProtoReflect() protoreflect.Message {
	mi := &file_zns_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Segment.ProtoReflect.Descriptor instead.
func (*Segment) Descriptor() ([]byte, []int) {
	return file_zns_proto_rawDescGZIP(), []int{0}
}

func (x *Segment) GetSegmentNo() int32 {
	if x != nil {
		return x.SegmentNo
	}
	return 0
}

func (x *Segment) GetMap() []byte {
	if x != nil {
		return x.Map
	}
	return nil
}

func (x *Segment) GetSegmentType() int32 {
	if x != nil {
		return x.SegmentType
	}
	return 0
}

type SegmentResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time    int64    `protobuf:"varint,1,opt,name=time,proto3" json:"time,omitempty"`
	ZoneNo  int32    `protobuf:"varint,2,opt,name=zone_no,json=zoneNo,proto3" json:"zone_no,omitempty"`
	Segment *Segment `protobuf:"bytes,3,opt,name=segment,proto3" json:"segment,omitempty"`
}

func (x *SegmentResponse) Reset() {
	*x = SegmentResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zns_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SegmentResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SegmentResponse) ProtoMessage() {}

func (x *SegmentResponse) ProtoReflect() protoreflect.Message {
	mi := &file_zns_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SegmentResponse.ProtoReflect.Descriptor instead.
func (*SegmentResponse) Descriptor() ([]byte, []int) {
	return file_zns_proto_rawDescGZIP(), []int{1}
}

func (x *SegmentResponse) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *SegmentResponse) GetZoneNo() int32 {
	if x != nil {
		return x.ZoneNo
	}
	return 0
}

func (x *SegmentResponse) GetSegment() *Segment {
	if x != nil {
		return x.Segment
	}
	return nil
}

type ZoneResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time   int64 `protobuf:"varint,1,opt,name=time,proto3" json:"time,omitempty"`
	ZoneNo int32 `protobuf:"varint,2,opt,name=zone_no,json=zoneNo,proto3" json:"zone_no,omitempty"`
	// -2: NotChanged, -1: Unknown, 0: HotData, 1: WarmData, 2: ColdData, 3: HotNode, 4: WarmNode, 5: ColdNode
	LastSegmentType int32      `protobuf:"varint,3,opt,name=last_segment_type,json=lastSegmentType,proto3" json:"last_segment_type,omitempty"`
	Segments        []*Segment `protobuf:"bytes,4,rep,name=segments,proto3" json:"segments,omitempty"`
}

func (x *ZoneResponse) Reset() {
	*x = ZoneResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_zns_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ZoneResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ZoneResponse) ProtoMessage() {}

func (x *ZoneResponse) ProtoReflect() protoreflect.Message {
	mi := &file_zns_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ZoneResponse.ProtoReflect.Descriptor instead.
func (*ZoneResponse) Descriptor() ([]byte, []int) {
	return file_zns_proto_rawDescGZIP(), []int{2}
}

func (x *ZoneResponse) GetTime() int64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *ZoneResponse) GetZoneNo() int32 {
	if x != nil {
		return x.ZoneNo
	}
	return 0
}

func (x *ZoneResponse) GetLastSegmentType() int32 {
	if x != nil {
		return x.LastSegmentType
	}
	return 0
}

func (x *ZoneResponse) GetSegments() []*Segment {
	if x != nil {
		return x.Segments
	}
	return nil
}

var File_zns_proto protoreflect.FileDescriptor

var file_zns_proto_rawDesc = []byte{
	0x0a, 0x09, 0x7a, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x5d, 0x0a, 0x07, 0x53,
	0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e,
	0x74, 0x5f, 0x6e, 0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x73, 0x65, 0x67, 0x6d,
	0x65, 0x6e, 0x74, 0x4e, 0x6f, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x61, 0x70, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x03, 0x6d, 0x61, 0x70, 0x12, 0x21, 0x0a, 0x0c, 0x73, 0x65, 0x67, 0x6d, 0x65,
	0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x73,
	0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x22, 0x62, 0x0a, 0x0f, 0x53, 0x65,
	0x67, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a,
	0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x74, 0x69, 0x6d,
	0x65, 0x12, 0x17, 0x0a, 0x07, 0x7a, 0x6f, 0x6e, 0x65, 0x5f, 0x6e, 0x6f, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x06, 0x7a, 0x6f, 0x6e, 0x65, 0x4e, 0x6f, 0x12, 0x22, 0x0a, 0x07, 0x73, 0x65,
	0x67, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x53, 0x65,
	0x67, 0x6d, 0x65, 0x6e, 0x74, 0x52, 0x07, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x22, 0x8d,
	0x01, 0x0a, 0x0c, 0x5a, 0x6f, 0x6e, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x74,
	0x69, 0x6d, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x7a, 0x6f, 0x6e, 0x65, 0x5f, 0x6e, 0x6f, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x7a, 0x6f, 0x6e, 0x65, 0x4e, 0x6f, 0x12, 0x2a, 0x0a, 0x11,
	0x6c, 0x61, 0x73, 0x74, 0x5f, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0f, 0x6c, 0x61, 0x73, 0x74, 0x53, 0x65, 0x67,
	0x6d, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x24, 0x0a, 0x08, 0x73, 0x65, 0x67, 0x6d,
	0x65, 0x6e, 0x74, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x53, 0x65, 0x67,
	0x6d, 0x65, 0x6e, 0x74, 0x52, 0x08, 0x73, 0x65, 0x67, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x42, 0x43,
	0x5a, 0x41, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x70, 0x69, 0x6e,
	0x67, 0x78, 0x69, 0x61, 0x6e, 0x67, 0x2d, 0x63, 0x68, 0x65, 0x6e, 0x2f, 0x62, 0x70, 0x66, 0x2d,
	0x66, 0x32, 0x66, 0x73, 0x2d, 0x7a, 0x6f, 0x6e, 0x65, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2f, 0x76,
	0x69, 0x65, 0x77, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x73, 0x3b, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_zns_proto_rawDescOnce sync.Once
	file_zns_proto_rawDescData = file_zns_proto_rawDesc
)

func file_zns_proto_rawDescGZIP() []byte {
	file_zns_proto_rawDescOnce.Do(func() {
		file_zns_proto_rawDescData = protoimpl.X.CompressGZIP(file_zns_proto_rawDescData)
	})
	return file_zns_proto_rawDescData
}

var file_zns_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_zns_proto_goTypes = []interface{}{
	(*Segment)(nil),         // 0: Segment
	(*SegmentResponse)(nil), // 1: SegmentResponse
	(*ZoneResponse)(nil),    // 2: ZoneResponse
}
var file_zns_proto_depIdxs = []int32{
	0, // 0: SegmentResponse.segment:type_name -> Segment
	0, // 1: ZoneResponse.segments:type_name -> Segment
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_zns_proto_init() }
func file_zns_proto_init() {
	if File_zns_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_zns_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Segment); i {
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
		file_zns_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SegmentResponse); i {
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
		file_zns_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ZoneResponse); i {
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
			RawDescriptor: file_zns_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_zns_proto_goTypes,
		DependencyIndexes: file_zns_proto_depIdxs,
		MessageInfos:      file_zns_proto_msgTypes,
	}.Build()
	File_zns_proto = out.File
	file_zns_proto_rawDesc = nil
	file_zns_proto_goTypes = nil
	file_zns_proto_depIdxs = nil
}
