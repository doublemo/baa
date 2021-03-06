// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.11.1
// source: internal/proto/pb/agent.proto

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
		mi := &file_internal_proto_pb_agent_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent) ProtoMessage() {}

func (x *Agent) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_agent_proto_msgTypes[0]
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
	return file_internal_proto_pb_agent_proto_rawDescGZIP(), []int{0}
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
		mi := &file_internal_proto_pb_agent_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Handshake) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Handshake) ProtoMessage() {}

func (x *Agent_Handshake) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_agent_proto_msgTypes[1]
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
	return file_internal_proto_pb_agent_proto_rawDescGZIP(), []int{0, 0}
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
		mi := &file_internal_proto_pb_agent_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Heartbeater) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Heartbeater) ProtoMessage() {}

func (x *Agent_Heartbeater) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_agent_proto_msgTypes[2]
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
	return file_internal_proto_pb_agent_proto_rawDescGZIP(), []int{0, 1}
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
		mi := &file_internal_proto_pb_agent_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Webrtc) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Webrtc) ProtoMessage() {}

func (x *Agent_Webrtc) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_agent_proto_msgTypes[3]
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
	return file_internal_proto_pb_agent_proto_rawDescGZIP(), []int{0, 2}
}

type Agent_KickedOut struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PeerID []string `protobuf:"bytes,1,rep,name=PeerID,proto3" json:"PeerID,omitempty"`
}

func (x *Agent_KickedOut) Reset() {
	*x = Agent_KickedOut{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_agent_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_KickedOut) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_KickedOut) ProtoMessage() {}

func (x *Agent_KickedOut) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_agent_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Agent_KickedOut.ProtoReflect.Descriptor instead.
func (*Agent_KickedOut) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_agent_proto_rawDescGZIP(), []int{0, 3}
}

func (x *Agent_KickedOut) GetPeerID() []string {
	if x != nil {
		return x.PeerID
	}
	return nil
}

type Agent_BroadcastMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Receiver   []uint64 `protobuf:"varint,1,rep,packed,name=Receiver,proto3" json:"Receiver,omitempty"`
	Command    int32    `protobuf:"varint,2,opt,name=Command,proto3" json:"Command,omitempty"`
	Payload    []byte   `protobuf:"bytes,3,opt,name=Payload,proto3" json:"Payload,omitempty"`
	SubCommand int32    `protobuf:"varint,4,opt,name=SubCommand,proto3" json:"SubCommand,omitempty"`
}

func (x *Agent_BroadcastMessage) Reset() {
	*x = Agent_BroadcastMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_agent_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_BroadcastMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_BroadcastMessage) ProtoMessage() {}

func (x *Agent_BroadcastMessage) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_agent_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Agent_BroadcastMessage.ProtoReflect.Descriptor instead.
func (*Agent_BroadcastMessage) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_agent_proto_rawDescGZIP(), []int{0, 4}
}

func (x *Agent_BroadcastMessage) GetReceiver() []uint64 {
	if x != nil {
		return x.Receiver
	}
	return nil
}

func (x *Agent_BroadcastMessage) GetCommand() int32 {
	if x != nil {
		return x.Command
	}
	return 0
}

func (x *Agent_BroadcastMessage) GetPayload() []byte {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *Agent_BroadcastMessage) GetSubCommand() int32 {
	if x != nil {
		return x.SubCommand
	}
	return 0
}

type Agent_Broadcast struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Messages []*Agent_BroadcastMessage `protobuf:"bytes,1,rep,name=Messages,proto3" json:"Messages,omitempty"`
}

func (x *Agent_Broadcast) Reset() {
	*x = Agent_Broadcast{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_agent_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Broadcast) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Broadcast) ProtoMessage() {}

func (x *Agent_Broadcast) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_agent_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Agent_Broadcast.ProtoReflect.Descriptor instead.
func (*Agent_Broadcast) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_agent_proto_rawDescGZIP(), []int{0, 5}
}

func (x *Agent_Broadcast) GetMessages() []*Agent_BroadcastMessage {
	if x != nil {
		return x.Messages
	}
	return nil
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
		mi := &file_internal_proto_pb_agent_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Webrtc_Trickle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Webrtc_Trickle) ProtoMessage() {}

func (x *Agent_Webrtc_Trickle) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_agent_proto_msgTypes[7]
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
	return file_internal_proto_pb_agent_proto_rawDescGZIP(), []int{0, 2, 0}
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
		mi := &file_internal_proto_pb_agent_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Agent_Webrtc_Signal) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Agent_Webrtc_Signal) ProtoMessage() {}

func (x *Agent_Webrtc_Signal) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_agent_proto_msgTypes[8]
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
	return file_internal_proto_pb_agent_proto_rawDescGZIP(), []int{0, 2, 1}
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

var File_internal_proto_pb_agent_proto protoreflect.FileDescriptor

var file_internal_proto_pb_agent_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x70, 0x62, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x02, 0x70, 0x62, 0x22, 0x96, 0x04, 0x0a, 0x05, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x1a, 0x2b, 0x0a,
	0x09, 0x48, 0x61, 0x6e, 0x64, 0x73, 0x68, 0x61, 0x6b, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x45, 0x31,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x45, 0x31, 0x12, 0x0e, 0x0a, 0x02, 0x45, 0x32,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x45, 0x32, 0x1a, 0x1b, 0x0a, 0x0b, 0x48, 0x65,
	0x61, 0x72, 0x74, 0x62, 0x65, 0x61, 0x74, 0x65, 0x72, 0x12, 0x0c, 0x0a, 0x01, 0x52, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x01, 0x52, 0x1a, 0xd3, 0x01, 0x0a, 0x06, 0x57, 0x65, 0x62, 0x72,
	0x74, 0x63, 0x1a, 0x27, 0x0a, 0x07, 0x54, 0x72, 0x69, 0x63, 0x6b, 0x6c, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x43, 0x61, 0x6e, 0x64, 0x69, 0x64, 0x61, 0x74, 0x65, 0x1a, 0x9f, 0x01, 0x0a, 0x06,
	0x53, 0x69, 0x67, 0x6e, 0x61, 0x6c, 0x12, 0x22, 0x0a, 0x0b, 0x44, 0x65, 0x73, 0x63, 0x72, 0x69,
	0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x48, 0x00, 0x52, 0x0b, 0x44,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x34, 0x0a, 0x07, 0x54, 0x72,
	0x69, 0x63, 0x6b, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x62,
	0x2e, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x57, 0x65, 0x62, 0x72, 0x74, 0x63, 0x2e, 0x54, 0x72,
	0x69, 0x63, 0x6b, 0x6c, 0x65, 0x48, 0x00, 0x52, 0x07, 0x54, 0x72, 0x69, 0x63, 0x6b, 0x6c, 0x65,
	0x12, 0x30, 0x0a, 0x12, 0x49, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x12,
	0x49, 0x63, 0x65, 0x43, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x42, 0x09, 0x0a, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x1a, 0x23, 0x0a,
	0x09, 0x4b, 0x69, 0x63, 0x6b, 0x65, 0x64, 0x4f, 0x75, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x50, 0x65,
	0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x06, 0x50, 0x65, 0x65, 0x72,
	0x49, 0x44, 0x1a, 0x82, 0x01, 0x0a, 0x10, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x52, 0x65, 0x63, 0x65, 0x69,
	0x76, 0x65, 0x72, 0x18, 0x01, 0x20, 0x03, 0x28, 0x04, 0x52, 0x08, 0x52, 0x65, 0x63, 0x65, 0x69,
	0x76, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x18, 0x0a,
	0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07,
	0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x1e, 0x0a, 0x0a, 0x53, 0x75, 0x62, 0x43, 0x6f,
	0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x53, 0x75, 0x62,
	0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x1a, 0x43, 0x0a, 0x09, 0x42, 0x72, 0x6f, 0x61, 0x64,
	0x63, 0x61, 0x73, 0x74, 0x12, 0x36, 0x0a, 0x08, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x70, 0x62, 0x2e, 0x41, 0x67, 0x65, 0x6e,
	0x74, 0x2e, 0x42, 0x72, 0x6f, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x52, 0x08, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x42, 0x2b, 0x5a, 0x29,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x6f, 0x75, 0x62, 0x6c,
	0x65, 0x6d, 0x6f, 0x2f, 0x62, 0x61, 0x61, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_internal_proto_pb_agent_proto_rawDescOnce sync.Once
	file_internal_proto_pb_agent_proto_rawDescData = file_internal_proto_pb_agent_proto_rawDesc
)

func file_internal_proto_pb_agent_proto_rawDescGZIP() []byte {
	file_internal_proto_pb_agent_proto_rawDescOnce.Do(func() {
		file_internal_proto_pb_agent_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_proto_pb_agent_proto_rawDescData)
	})
	return file_internal_proto_pb_agent_proto_rawDescData
}

var file_internal_proto_pb_agent_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_internal_proto_pb_agent_proto_goTypes = []interface{}{
	(*Agent)(nil),                  // 0: pb.Agent
	(*Agent_Handshake)(nil),        // 1: pb.Agent.Handshake
	(*Agent_Heartbeater)(nil),      // 2: pb.Agent.Heartbeater
	(*Agent_Webrtc)(nil),           // 3: pb.Agent.Webrtc
	(*Agent_KickedOut)(nil),        // 4: pb.Agent.KickedOut
	(*Agent_BroadcastMessage)(nil), // 5: pb.Agent.BroadcastMessage
	(*Agent_Broadcast)(nil),        // 6: pb.Agent.Broadcast
	(*Agent_Webrtc_Trickle)(nil),   // 7: pb.Agent.Webrtc.Trickle
	(*Agent_Webrtc_Signal)(nil),    // 8: pb.Agent.Webrtc.Signal
}
var file_internal_proto_pb_agent_proto_depIdxs = []int32{
	5, // 0: pb.Agent.Broadcast.Messages:type_name -> pb.Agent.BroadcastMessage
	7, // 1: pb.Agent.Webrtc.Signal.Trickle:type_name -> pb.Agent.Webrtc.Trickle
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_internal_proto_pb_agent_proto_init() }
func file_internal_proto_pb_agent_proto_init() {
	if File_internal_proto_pb_agent_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_proto_pb_agent_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
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
		file_internal_proto_pb_agent_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
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
		file_internal_proto_pb_agent_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
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
		file_internal_proto_pb_agent_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
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
		file_internal_proto_pb_agent_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Agent_KickedOut); i {
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
		file_internal_proto_pb_agent_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Agent_BroadcastMessage); i {
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
		file_internal_proto_pb_agent_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Agent_Broadcast); i {
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
		file_internal_proto_pb_agent_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
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
		file_internal_proto_pb_agent_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
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
	file_internal_proto_pb_agent_proto_msgTypes[8].OneofWrappers = []interface{}{
		(*Agent_Webrtc_Signal_Description)(nil),
		(*Agent_Webrtc_Signal_Trickle)(nil),
		(*Agent_Webrtc_Signal_IceConnectionState)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_proto_pb_agent_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_proto_pb_agent_proto_goTypes,
		DependencyIndexes: file_internal_proto_pb_agent_proto_depIdxs,
		MessageInfos:      file_internal_proto_pb_agent_proto_msgTypes,
	}.Build()
	File_internal_proto_pb_agent_proto = out.File
	file_internal_proto_pb_agent_proto_rawDesc = nil
	file_internal_proto_pb_agent_proto_goTypes = nil
	file_internal_proto_pb_agent_proto_depIdxs = nil
}
