// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.11.1
// source: internal/proto/pb/robot.proto

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

type Robot struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Robot) Reset() {
	*x = Robot{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot) ProtoMessage() {}

func (x *Robot) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot.ProtoReflect.Descriptor instead.
func (*Robot) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0}
}

type Robot_Info struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID        uint64 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	AccountID uint64 `protobuf:"varint,2,opt,name=AccountID,proto3" json:"AccountID,omitempty"`
	UnionID   uint64 `protobuf:"varint,3,opt,name=UnionID,proto3" json:"UnionID,omitempty"`
	UserID    uint64 `protobuf:"varint,4,opt,name=UserID,proto3" json:"UserID,omitempty"`
	Nickname  string `protobuf:"bytes,5,opt,name=Nickname,proto3" json:"Nickname,omitempty"`
	Headimg   string `protobuf:"bytes,6,opt,name=Headimg,proto3" json:"Headimg,omitempty"`
	Age       int32  `protobuf:"varint,7,opt,name=Age,proto3" json:"Age,omitempty"`
	Sex       int32  `protobuf:"varint,8,opt,name=Sex,proto3" json:"Sex,omitempty"`
	Idcard    string `protobuf:"bytes,9,opt,name=Idcard,proto3" json:"Idcard,omitempty"`
	Phone     string `protobuf:"bytes,10,opt,name=Phone,proto3" json:"Phone,omitempty"`
}

func (x *Robot_Info) Reset() {
	*x = Robot_Info{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Info) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Info) ProtoMessage() {}

func (x *Robot_Info) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Info.ProtoReflect.Descriptor instead.
func (*Robot_Info) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 0}
}

func (x *Robot_Info) GetID() uint64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *Robot_Info) GetAccountID() uint64 {
	if x != nil {
		return x.AccountID
	}
	return 0
}

func (x *Robot_Info) GetUnionID() uint64 {
	if x != nil {
		return x.UnionID
	}
	return 0
}

func (x *Robot_Info) GetUserID() uint64 {
	if x != nil {
		return x.UserID
	}
	return 0
}

func (x *Robot_Info) GetNickname() string {
	if x != nil {
		return x.Nickname
	}
	return ""
}

func (x *Robot_Info) GetHeadimg() string {
	if x != nil {
		return x.Headimg
	}
	return ""
}

func (x *Robot_Info) GetAge() int32 {
	if x != nil {
		return x.Age
	}
	return 0
}

func (x *Robot_Info) GetSex() int32 {
	if x != nil {
		return x.Sex
	}
	return 0
}

func (x *Robot_Info) GetIdcard() string {
	if x != nil {
		return x.Idcard
	}
	return ""
}

func (x *Robot_Info) GetPhone() string {
	if x != nil {
		return x.Phone
	}
	return ""
}

type Robot_Create struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Robot_Create) Reset() {
	*x = Robot_Create{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Create) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Create) ProtoMessage() {}

func (x *Robot_Create) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Create.ProtoReflect.Descriptor instead.
func (*Robot_Create) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 1}
}

type Robot_Start struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Robot_Start) Reset() {
	*x = Robot_Start{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Start) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Start) ProtoMessage() {}

func (x *Robot_Start) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Start.ProtoReflect.Descriptor instead.
func (*Robot_Start) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 2}
}

type Robot_Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *Robot_Status) Reset() {
	*x = Robot_Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Status) ProtoMessage() {}

func (x *Robot_Status) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Status.ProtoReflect.Descriptor instead.
func (*Robot_Status) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 3}
}

type Robot_Create_Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Payload:
	//	*Robot_Create_Request_Account
	//	*Robot_Create_Request_Register
	Payload isRobot_Create_Request_Payload `protobuf_oneof:"Payload"`
}

func (x *Robot_Create_Request) Reset() {
	*x = Robot_Create_Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Create_Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Create_Request) ProtoMessage() {}

func (x *Robot_Create_Request) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Create_Request.ProtoReflect.Descriptor instead.
func (*Robot_Create_Request) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 1, 0}
}

func (m *Robot_Create_Request) GetPayload() isRobot_Create_Request_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *Robot_Create_Request) GetAccount() *Robot_Create_Account {
	if x, ok := x.GetPayload().(*Robot_Create_Request_Account); ok {
		return x.Account
	}
	return nil
}

func (x *Robot_Create_Request) GetRegister() *Robot_Create_Register {
	if x, ok := x.GetPayload().(*Robot_Create_Request_Register); ok {
		return x.Register
	}
	return nil
}

type isRobot_Create_Request_Payload interface {
	isRobot_Create_Request_Payload()
}

type Robot_Create_Request_Account struct {
	Account *Robot_Create_Account `protobuf:"bytes,1,opt,name=Account,proto3,oneof"`
}

type Robot_Create_Request_Register struct {
	Register *Robot_Create_Register `protobuf:"bytes,2,opt,name=Register,proto3,oneof"`
}

func (*Robot_Create_Request_Account) isRobot_Create_Request_Payload() {}

func (*Robot_Create_Request_Register) isRobot_Create_Request_Payload() {}

type Robot_Create_Reply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OK bool `protobuf:"varint,1,opt,name=OK,proto3" json:"OK,omitempty"`
}

func (x *Robot_Create_Reply) Reset() {
	*x = Robot_Create_Reply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Create_Reply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Create_Reply) ProtoMessage() {}

func (x *Robot_Create_Reply) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Create_Reply.ProtoReflect.Descriptor instead.
func (*Robot_Create_Reply) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 1, 1}
}

func (x *Robot_Create_Reply) GetOK() bool {
	if x != nil {
		return x.OK
	}
	return false
}

type Robot_Create_Account struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name   string `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	Secret string `protobuf:"bytes,2,opt,name=Secret,proto3" json:"Secret,omitempty"`
}

func (x *Robot_Create_Account) Reset() {
	*x = Robot_Create_Account{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Create_Account) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Create_Account) ProtoMessage() {}

func (x *Robot_Create_Account) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Create_Account.ProtoReflect.Descriptor instead.
func (*Robot_Create_Account) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 1, 2}
}

func (x *Robot_Create_Account) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Robot_Create_Account) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

type Robot_Create_Register struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Schema   string `protobuf:"bytes,1,opt,name=Schema,proto3" json:"Schema,omitempty"`
	Name     string `protobuf:"bytes,2,opt,name=Name,proto3" json:"Name,omitempty"`
	Secret   string `protobuf:"bytes,3,opt,name=Secret,proto3" json:"Secret,omitempty"`
	Nickname string `protobuf:"bytes,4,opt,name=Nickname,proto3" json:"Nickname,omitempty"`
	Headimg  string `protobuf:"bytes,5,opt,name=Headimg,proto3" json:"Headimg,omitempty"`
	Age      int32  `protobuf:"varint,6,opt,name=Age,proto3" json:"Age,omitempty"`
	Sex      int32  `protobuf:"varint,7,opt,name=Sex,proto3" json:"Sex,omitempty"`
	Idcard   string `protobuf:"bytes,8,opt,name=Idcard,proto3" json:"Idcard,omitempty"`
	Phone    string `protobuf:"bytes,9,opt,name=Phone,proto3" json:"Phone,omitempty"`
}

func (x *Robot_Create_Register) Reset() {
	*x = Robot_Create_Register{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Create_Register) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Create_Register) ProtoMessage() {}

func (x *Robot_Create_Register) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Create_Register.ProtoReflect.Descriptor instead.
func (*Robot_Create_Register) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 1, 3}
}

func (x *Robot_Create_Register) GetSchema() string {
	if x != nil {
		return x.Schema
	}
	return ""
}

func (x *Robot_Create_Register) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Robot_Create_Register) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

func (x *Robot_Create_Register) GetNickname() string {
	if x != nil {
		return x.Nickname
	}
	return ""
}

func (x *Robot_Create_Register) GetHeadimg() string {
	if x != nil {
		return x.Headimg
	}
	return ""
}

func (x *Robot_Create_Register) GetAge() int32 {
	if x != nil {
		return x.Age
	}
	return 0
}

func (x *Robot_Create_Register) GetSex() int32 {
	if x != nil {
		return x.Sex
	}
	return 0
}

func (x *Robot_Create_Register) GetIdcard() string {
	if x != nil {
		return x.Idcard
	}
	return ""
}

func (x *Robot_Create_Register) GetPhone() string {
	if x != nil {
		return x.Phone
	}
	return ""
}

type Robot_Start_Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Values []uint64 `protobuf:"varint,1,rep,packed,name=Values,proto3" json:"Values,omitempty"`
}

func (x *Robot_Start_Request) Reset() {
	*x = Robot_Start_Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Start_Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Start_Request) ProtoMessage() {}

func (x *Robot_Start_Request) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Start_Request.ProtoReflect.Descriptor instead.
func (*Robot_Start_Request) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 2, 0}
}

func (x *Robot_Start_Request) GetValues() []uint64 {
	if x != nil {
		return x.Values
	}
	return nil
}

type Robot_Start_Reply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Succeeded []uint64 `protobuf:"varint,1,rep,packed,name=Succeeded,proto3" json:"Succeeded,omitempty"`
	Failed    []uint64 `protobuf:"varint,2,rep,packed,name=Failed,proto3" json:"Failed,omitempty"`
}

func (x *Robot_Start_Reply) Reset() {
	*x = Robot_Start_Reply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Start_Reply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Start_Reply) ProtoMessage() {}

func (x *Robot_Start_Reply) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Start_Reply.ProtoReflect.Descriptor instead.
func (*Robot_Start_Reply) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 2, 1}
}

func (x *Robot_Start_Reply) GetSucceeded() []uint64 {
	if x != nil {
		return x.Succeeded
	}
	return nil
}

func (x *Robot_Start_Reply) GetFailed() []uint64 {
	if x != nil {
		return x.Failed
	}
	return nil
}

type Robot_Status_Info struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID       uint64 `protobuf:"varint,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Nickname string `protobuf:"bytes,2,opt,name=Nickname,proto3" json:"Nickname,omitempty"`
	Headimg  string `protobuf:"bytes,3,opt,name=Headimg,proto3" json:"Headimg,omitempty"`
	Age      int32  `protobuf:"varint,4,opt,name=Age,proto3" json:"Age,omitempty"`
	Sex      int32  `protobuf:"varint,5,opt,name=Sex,proto3" json:"Sex,omitempty"`
}

func (x *Robot_Status_Info) Reset() {
	*x = Robot_Status_Info{}
	if protoimpl.UnsafeEnabled {
		mi := &file_internal_proto_pb_robot_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Robot_Status_Info) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Robot_Status_Info) ProtoMessage() {}

func (x *Robot_Status_Info) ProtoReflect() protoreflect.Message {
	mi := &file_internal_proto_pb_robot_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Robot_Status_Info.ProtoReflect.Descriptor instead.
func (*Robot_Status_Info) Descriptor() ([]byte, []int) {
	return file_internal_proto_pb_robot_proto_rawDescGZIP(), []int{0, 3, 0}
}

func (x *Robot_Status_Info) GetID() uint64 {
	if x != nil {
		return x.ID
	}
	return 0
}

func (x *Robot_Status_Info) GetNickname() string {
	if x != nil {
		return x.Nickname
	}
	return ""
}

func (x *Robot_Status_Info) GetHeadimg() string {
	if x != nil {
		return x.Headimg
	}
	return ""
}

func (x *Robot_Status_Info) GetAge() int32 {
	if x != nil {
		return x.Age
	}
	return 0
}

func (x *Robot_Status_Info) GetSex() int32 {
	if x != nil {
		return x.Sex
	}
	return 0
}

var File_internal_proto_pb_robot_proto protoreflect.FileDescriptor

var file_internal_proto_pb_robot_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x70, 0x62, 0x2f, 0x72, 0x6f, 0x62, 0x6f, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x02, 0x70, 0x62, 0x22, 0x99, 0x07, 0x0a, 0x05, 0x52, 0x6f, 0x62, 0x6f, 0x74, 0x1a, 0xee, 0x01,
	0x0a, 0x04, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x04, 0x52, 0x02, 0x49, 0x44, 0x12, 0x1c, 0x0a, 0x09, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x49, 0x44, 0x18, 0x02, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x41, 0x63, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x49, 0x44, 0x12, 0x18, 0x0a, 0x07, 0x55, 0x6e, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x04, 0x52, 0x07, 0x55, 0x6e, 0x69, 0x6f, 0x6e, 0x49, 0x44, 0x12, 0x16,
	0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06,
	0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x08, 0x4e, 0x69, 0x63, 0x6b, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x4e, 0x69, 0x63, 0x6b, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x48, 0x65, 0x61, 0x64, 0x69, 0x6d, 0x67, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x48, 0x65, 0x61, 0x64, 0x69, 0x6d, 0x67, 0x12, 0x10, 0x0a, 0x03,
	0x41, 0x67, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x41, 0x67, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x53, 0x65, 0x78, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x53, 0x65, 0x78,
	0x12, 0x16, 0x0a, 0x06, 0x49, 0x64, 0x63, 0x61, 0x72, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x49, 0x64, 0x63, 0x61, 0x72, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x50, 0x68, 0x6f, 0x6e,
	0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x50, 0x68, 0x6f, 0x6e, 0x65, 0x1a, 0xb7,
	0x03, 0x0a, 0x06, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x1a, 0x83, 0x01, 0x0a, 0x07, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x34, 0x0a, 0x07, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x70, 0x62, 0x2e, 0x52, 0x6f, 0x62, 0x6f,
	0x74, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x2e, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x48, 0x00, 0x52, 0x07, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x37, 0x0a, 0x08, 0x52,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e,
	0x70, 0x62, 0x2e, 0x52, 0x6f, 0x62, 0x6f, 0x74, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x2e,
	0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x48, 0x00, 0x52, 0x08, 0x52, 0x65, 0x67, 0x69,
	0x73, 0x74, 0x65, 0x72, 0x42, 0x09, 0x0a, 0x07, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x1a,
	0x17, 0x0a, 0x05, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x4f, 0x4b, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x4f, 0x4b, 0x1a, 0x35, 0x0a, 0x07, 0x41, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x53, 0x65, 0x63, 0x72, 0x65,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x1a,
	0xd6, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06,
	0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x53, 0x63,
	0x68, 0x65, 0x6d, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x53, 0x65, 0x63, 0x72,
	0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74,
	0x12, 0x1a, 0x0a, 0x08, 0x4e, 0x69, 0x63, 0x6b, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x4e, 0x69, 0x63, 0x6b, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x48, 0x65, 0x61, 0x64, 0x69, 0x6d, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x48,
	0x65, 0x61, 0x64, 0x69, 0x6d, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x41, 0x67, 0x65, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x03, 0x41, 0x67, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x53, 0x65, 0x78, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x53, 0x65, 0x78, 0x12, 0x16, 0x0a, 0x06, 0x49, 0x64,
	0x63, 0x61, 0x72, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x49, 0x64, 0x63, 0x61,
	0x72, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x50, 0x68, 0x6f, 0x6e, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x50, 0x68, 0x6f, 0x6e, 0x65, 0x1a, 0x69, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x72,
	0x74, 0x1a, 0x21, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x56, 0x61, 0x6c, 0x75, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x04, 0x52, 0x06, 0x56, 0x61,
	0x6c, 0x75, 0x65, 0x73, 0x1a, 0x3d, 0x0a, 0x05, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x1c, 0x0a,
	0x09, 0x53, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x18, 0x01, 0x20, 0x03, 0x28, 0x04,
	0x52, 0x09, 0x53, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64, 0x65, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x46,
	0x61, 0x69, 0x6c, 0x65, 0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x04, 0x52, 0x06, 0x46, 0x61, 0x69,
	0x6c, 0x65, 0x64, 0x1a, 0x7a, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x1a, 0x70, 0x0a,
	0x04, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x02, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x08, 0x4e, 0x69, 0x63, 0x6b, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x4e, 0x69, 0x63, 0x6b, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x48, 0x65, 0x61, 0x64, 0x69, 0x6d, 0x67, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x48, 0x65, 0x61, 0x64, 0x69, 0x6d, 0x67, 0x12, 0x10, 0x0a, 0x03, 0x41,
	0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x41, 0x67, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x53, 0x65, 0x78, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x53, 0x65, 0x78, 0x42,
	0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x64, 0x6f,
	0x75, 0x62, 0x6c, 0x65, 0x6d, 0x6f, 0x2f, 0x62, 0x61, 0x61, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_internal_proto_pb_robot_proto_rawDescOnce sync.Once
	file_internal_proto_pb_robot_proto_rawDescData = file_internal_proto_pb_robot_proto_rawDesc
)

func file_internal_proto_pb_robot_proto_rawDescGZIP() []byte {
	file_internal_proto_pb_robot_proto_rawDescOnce.Do(func() {
		file_internal_proto_pb_robot_proto_rawDescData = protoimpl.X.CompressGZIP(file_internal_proto_pb_robot_proto_rawDescData)
	})
	return file_internal_proto_pb_robot_proto_rawDescData
}

var file_internal_proto_pb_robot_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_internal_proto_pb_robot_proto_goTypes = []interface{}{
	(*Robot)(nil),                 // 0: pb.Robot
	(*Robot_Info)(nil),            // 1: pb.Robot.Info
	(*Robot_Create)(nil),          // 2: pb.Robot.Create
	(*Robot_Start)(nil),           // 3: pb.Robot.Start
	(*Robot_Status)(nil),          // 4: pb.Robot.Status
	(*Robot_Create_Request)(nil),  // 5: pb.Robot.Create.Request
	(*Robot_Create_Reply)(nil),    // 6: pb.Robot.Create.Reply
	(*Robot_Create_Account)(nil),  // 7: pb.Robot.Create.Account
	(*Robot_Create_Register)(nil), // 8: pb.Robot.Create.Register
	(*Robot_Start_Request)(nil),   // 9: pb.Robot.Start.Request
	(*Robot_Start_Reply)(nil),     // 10: pb.Robot.Start.Reply
	(*Robot_Status_Info)(nil),     // 11: pb.Robot.Status.Info
}
var file_internal_proto_pb_robot_proto_depIdxs = []int32{
	7, // 0: pb.Robot.Create.Request.Account:type_name -> pb.Robot.Create.Account
	8, // 1: pb.Robot.Create.Request.Register:type_name -> pb.Robot.Create.Register
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_internal_proto_pb_robot_proto_init() }
func file_internal_proto_pb_robot_proto_init() {
	if File_internal_proto_pb_robot_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_internal_proto_pb_robot_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Info); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Create); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Start); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Status); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Create_Request); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Create_Reply); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Create_Account); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Create_Register); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Start_Request); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Start_Reply); i {
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
		file_internal_proto_pb_robot_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Robot_Status_Info); i {
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
	file_internal_proto_pb_robot_proto_msgTypes[5].OneofWrappers = []interface{}{
		(*Robot_Create_Request_Account)(nil),
		(*Robot_Create_Request_Register)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_internal_proto_pb_robot_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_internal_proto_pb_robot_proto_goTypes,
		DependencyIndexes: file_internal_proto_pb_robot_proto_depIdxs,
		MessageInfos:      file_internal_proto_pb_robot_proto_msgTypes,
	}.Build()
	File_internal_proto_pb_robot_proto = out.File
	file_internal_proto_pb_robot_proto_rawDesc = nil
	file_internal_proto_pb_robot_proto_goTypes = nil
	file_internal_proto_pb_robot_proto_depIdxs = nil
}
