// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.11.1
// source: kits/agent/proto/pb/agent.proto

package pb

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Agent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Agent) Reset() {
	*x = Agent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent) ProtoMessage() {}

func (x *Agent) ProtoReflect() protoreflect.Message {
	mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Agent.ProtoReflect.Descriptor instead.
func (*Agent) Descriptor() ([]byte, []int) {
	return file_kits_agent_proto_pb_agent_proto_rawDescGZIP(), []int{0}
}

type Agent_Handshake struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	E1 int64 `protobuf:"varint,1,opt,name=E1,proto3" json:"E1,omitempty"`
	E2 int64 `protobuf:"varint,2,opt,name=E2,proto3" json:"E2,omitempty"`
}

func (x *Agent_Handshake) Reset() {
	*x = Agent_Handshake{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Handshake) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Handshake) ProtoMessage() {}

func (x *Agent_Handshake) ProtoReflect() protoreflect.Message {
	mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Agent_Handshake.ProtoReflect.Descriptor instead.
func (*Agent_Handshake) Descriptor() ([]byte, []int) {
	return file_kits_agent_proto_pb_agent_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Agent_Handshake) GetE1() int64 {
	if x != nil {
		return x.E1
	}
	return 0
}

func (x *Agent_Handshake) GetE2() int64 {
	if x != nil {
		return x.E2
	}
	return 0
}

type Agent_Heartbeater struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	R int64 `protobuf:"varint,1,opt,name=R,proto3" json:"R,omitempty"`
}

func (x *Agent_Heartbeater) Reset() {
	*x = Agent_Heartbeater{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Heartbeater) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Heartbeater) ProtoMessage() {}

func (x *Agent_Heartbeater) ProtoReflect() protoreflect.Message {
	mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Agent_Heartbeater.ProtoReflect.Descriptor instead.
func (*Agent_Heartbeater) Descriptor() ([]byte, []int) {
	return file_kits_agent_proto_pb_agent_proto_rawDescGZIP(), []int{0, 1}
}

func (x *Agent_Heartbeater) GetR() int64 {
	if x != nil {
		return x.R
	}
	return 0
}

type Agent_Webrtc struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Agent_Webrtc) Reset() {
	*x = Agent_Webrtc{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Webrtc) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Webrtc) ProtoMessage() {}

func (x *Agent_Webrtc) ProtoReflect() protoreflect.Message {
	mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Agent_Webrtc.ProtoReflect.Descriptor instead.
func (*Agent_Webrtc) Descriptor() ([]byte, []int) {
	return file_kits_agent_proto_pb_agent_proto_rawDescGZIP(), []int{0, 2}
}

type Agent_Webrtc_Trickle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Candidate string `protobuf:"bytes,1,opt,name=Candidate,proto3" json:"Candidate,omitempty"`
}

func (x *Agent_Webrtc_Trickle) Reset() {
	*x = Agent_Webrtc_Trickle{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Webrtc_Trickle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Webrtc_Trickle) ProtoMessage() {}

func (x *Agent_Webrtc_Trickle) ProtoReflect() protoreflect.Message {
	mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Agent_Webrtc_Trickle.ProtoReflect.Descriptor instead.
func (*Agent_Webrtc_Trickle) Descriptor() ([]byte, []int) {
	return file_kits_agent_proto_pb_agent_proto_rawDescGZIP(), []int{0, 2, 0}
}

func (x *Agent_Webrtc_Trickle) GetCandidate() string {
	if x != nil {
		return x.Candidate
	}
	return ""
}

type Agent_Webrtc_Signal struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Payload:
	//	*Agent_Webrtc_Signal_Description
	//	*Agent_Webrtc_Signal_Trickle
	//	*Agent_Webrtc_Signal_IceConnectionState
	Payload isAgent_Webrtc_Signal_Payload `protobuf_oneof:"Payload"`
}

func (x *Agent_Webrtc_Signal) Reset() {
	*x = Agent_Webrtc_Signal{}
	if protoimpl.UnsafeEnabled {
		mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Webrtc_Signal) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Webrtc_Signal) ProtoMessage() {}

func (x *Agent_Webrtc_Signal) ProtoReflect() protoreflect.Message {
	mi := &file_kits_agent_proto_pb_agent_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Agent_Webrtc_Signal.ProtoReflect.Descriptor instead.
func (*Agent_Webrtc_Signal) Descriptor() ([]byte, []int) {
	return file_kits_agent_proto_pb_agent_proto_rawDescGZIP(), []int{0, 2, 1}
}

func (m *Agent_Webrtc_Signal) GetPayload() isAgent_Webrtc_Signal_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *Agent_Webrtc_Signal) GetDescription() []byte {
	if x, ok := x.GetPayload().(*Agent_Webrtc_Signal_Description); ok {
		return x.Description
	}
	return nil
}

func (x *Agent_Webrtc_Signal) GetTrickle() *Agent_Webrtc_Trickle {
	if x, ok := x.GetPayload().(*Agent_Webrtc_Signal_Trickle); ok {
		return x.Trickle
	}
	return nil
}

func (x *Agent_Webrtc_Signal) GetIceConnectionState() string {
	if x, ok := x.GetPayload().(*Agent_Webrtc_Signal_IceConnectionState); ok {
		return x.IceConnectionState
	}
	return ""
}

type isAgent_Webrtc_Signal_Payload interface {
	isAgent_Webrtc_Signal_Payload()
}

type Agent_Webrtc_Signal_Description struct {
	Description []byte `protobuf:"bytes,1,opt,name=Description,proto3,oneof"`
}

type Agent_Webrtc_Signal_Trickle struct {
	Trickle *Agent_Webrtc_Trickle `protobuf:"bytes,2,opt,name=Trickle,proto3,oneof"`
}

type Agent_Webrtc_Signal_IceConnectionState struct {
	IceConnectionState string `protobuf:"bytes,3,opt,name=IceConnectionState,proto3,oneof"`
}

func (*Agent_Webrtc_Signal_Description) isAgent_Webrtc_Signal_Payload() {}

func (*Agent_Webrtc_Signal_Trickle) isAgent_Webrtc_Signal_Payload() {}

func (*Agent_Webrtc_Signal_IceConnectionState) isAgent_Webrtc_Signal_Payload() {}

var File_kits_agent_proto_pb_agent_proto protoreflect.FileDescriptor

var file_kits_agent_proto_pb_agent_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x6b, 0x69, 0x74, 0x73, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x70, 0x62, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x02, 0x70, 0x62, 0x22, 0xa7, 0x02, 0x0a, 0x05, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x1a,
	0x2b, 0x0a, 0x09, 0x48, 0x61, 0x6e, 0x64, 0x73, 0x68, 0x61, 0x6b, 0x65, 0x12, 0x0e, 0x0a, 0x02,
	0x45, 0x31, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x45, 0x31, 0x12, 0x0e, 0x0a, 0x02,
	0x45, 0x32, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x45, 0x32, 0x1a, 0x1b, 0x0a, 0x0b,
	0x48, 0x65, 0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x65, 0x72, 0x12, 0x0c, 0x0a, 0x01, 0x52,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x01, 0x52, 0x1a, 0xd3, 0x01, 0x0a, 0x06, 0x57, 0x65,
	0x62, 0x72, 0x74, 0x63, 0x1a, 0x27, 0x0a, 0x07, 0x54, 0x72, 0x69, 0x63, 0x6b, 0x6c, 0x65, 0x12,
	0x1c, 0x0a, 0x09, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x09, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x1a, 0x9f, 0x01,
	0x0a, 0x06, 0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x12, 0x22, 0x0a, 0x0b, 0x44, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52,
	0x0b, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x34, 0x0a, 0x07,
	0x54, 0x72, 0x69, 0x63, 0x6b, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e,
	0x70, 0x62, 0x2e, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x57, 0x65, 0x62, 0x72, 0x74, 0x63, 0x2e,
	0x54, 0x72, 0x69, 0x63, 0x6b, 0x6c, 0x65, 0x48, 0x00, 0x52, 0x07, 0x54, 0x72, 0x69, 0x63, 0x6b,
	0x6c, 0x65, 0x12, 0x30, 0x0a, 0x12, 0x49, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00,
	0x52, 0x12, 0x49, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x42, 0x09, 0x0a, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x42,
	0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x6f,
	0x75, 0x62, 0x6c, 0x65, 0x6d, 0x6f, 0x2f, 0x62, 0x61, 0x61, 0x2f, 0x6b, 0x69, 0x74, 0x73, 0x2f,
	0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x62, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_kits_agent_proto_pb_agent_proto_rawDescOnce sync.Once
	file_kits_agent_proto_pb_agent_proto_rawDescData = file_kits_agent_proto_pb_agent_proto_rawDesc
)

func file_kits_agent_proto_pb_agent_proto_rawDescGZIP() []byte {
	file_kits_agent_proto_pb_agent_proto_rawDescOnce.Do(func() {
		file_kits_agent_proto_pb_agent_proto_rawDescData = protoimpl.X.CompressGZIP(file_kits_agent_proto_pb_agent_proto_rawDescData)
	})
	return file_kits_agent_proto_pb_agent_proto_rawDescData
}

var file_kits_agent_proto_pb_agent_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_kits_agent_proto_pb_agent_proto_goTypes = []interface{}{
	(*Agent)(nil),                // 0: pb.Agent
	(*Agent_Handshake)(nil),      // 1: pb.Agent.Handshake
	(*Agent_Heartbeater)(nil),    // 2: pb.Agent.Heartbeater
	(*Agent_Webrtc)(nil),         // 3: pb.Agent.Webrtc
	(*Agent_Webrtc_Trickle)(nil), // 4: pb.Agent.Webrtc.Trickle
	(*Agent_Webrtc_Signal)(nil),  // 5: pb.Agent.Webrtc.Signal
}
var file_kits_agent_proto_pb_agent_proto_depIdxs = []int32{
	4, // 0: pb.Agent.Webrtc.Signal.Trickle:type_name -> pb.Agent.Webrtc.Trickle
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_kits_agent_proto_pb_agent_proto_init() }
func file_kits_agent_proto_pb_agent_proto_init() {
	if File_kits_agent_proto_pb_agent_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_kits_agent_proto_pb_agent_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Agent); i {
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
		file_kits_agent_proto_pb_agent_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Agent_Handshake); i {
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
		file_kits_agent_proto_pb_agent_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Agent_Heartbeater); i {
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
		file_kits_agent_proto_pb_agent_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Agent_Webrtc); i {
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
		file_kits_agent_proto_pb_agent_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Agent_Webrtc_Trickle); i {
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
		file_kits_agent_proto_pb_agent_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Agent_Webrtc_Signal); i {
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
	file_kits_agent_proto_pb_agent_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*Agent_Webrtc_Signal_Description)(nil),
		(*Agent_Webrtc_Signal_Trickle)(nil),
		(*Agent_Webrtc_Signal_IceConnectionState)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_kits_agent_proto_pb_agent_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_kits_agent_proto_pb_agent_proto_goTypes,
		DependencyIndexes: file_kits_agent_proto_pb_agent_proto_depIdxs,
		MessageInfos:      file_kits_agent_proto_pb_agent_proto_msgTypes,
	}.Build()
	File_kits_agent_proto_pb_agent_proto = out.File
	file_kits_agent_proto_pb_agent_proto_rawDesc = nil
	file_kits_agent_proto_pb_agent_proto_goTypes = nil
	file_kits_agent_proto_pb_agent_proto_depIdxs = nil
}
