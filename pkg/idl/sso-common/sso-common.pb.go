// Code generated by protoc-gen-go. DO NOT EDIT.
// source: sso-common/sso-common.proto

package sso_common

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ResponseStatus struct {
	Errors               []*ResponseStatus_Error `protobuf:"bytes,1,rep,name=errors,proto3" json:"errors,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *ResponseStatus) Reset()         { *m = ResponseStatus{} }
func (m *ResponseStatus) String() string { return proto.CompactTextString(m) }
func (*ResponseStatus) ProtoMessage()    {}
func (*ResponseStatus) Descriptor() ([]byte, []int) {
	return fileDescriptor_cc8776310bb7bdb9, []int{0}
}

func (m *ResponseStatus) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResponseStatus.Unmarshal(m, b)
}
func (m *ResponseStatus) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResponseStatus.Marshal(b, m, deterministic)
}
func (m *ResponseStatus) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResponseStatus.Merge(m, src)
}
func (m *ResponseStatus) XXX_Size() int {
	return xxx_messageInfo_ResponseStatus.Size(m)
}
func (m *ResponseStatus) XXX_DiscardUnknown() {
	xxx_messageInfo_ResponseStatus.DiscardUnknown(m)
}

var xxx_messageInfo_ResponseStatus proto.InternalMessageInfo

func (m *ResponseStatus) GetErrors() []*ResponseStatus_Error {
	if m != nil {
		return m.Errors
	}
	return nil
}

type ResponseStatus_Error struct {
	Slug                 string   `protobuf:"bytes,1,opt,name=slug,proto3" json:"slug,omitempty"`
	Message              string   `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ResponseStatus_Error) Reset()         { *m = ResponseStatus_Error{} }
func (m *ResponseStatus_Error) String() string { return proto.CompactTextString(m) }
func (*ResponseStatus_Error) ProtoMessage()    {}
func (*ResponseStatus_Error) Descriptor() ([]byte, []int) {
	return fileDescriptor_cc8776310bb7bdb9, []int{0, 0}
}

func (m *ResponseStatus_Error) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResponseStatus_Error.Unmarshal(m, b)
}
func (m *ResponseStatus_Error) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResponseStatus_Error.Marshal(b, m, deterministic)
}
func (m *ResponseStatus_Error) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResponseStatus_Error.Merge(m, src)
}
func (m *ResponseStatus_Error) XXX_Size() int {
	return xxx_messageInfo_ResponseStatus_Error.Size(m)
}
func (m *ResponseStatus_Error) XXX_DiscardUnknown() {
	xxx_messageInfo_ResponseStatus_Error.DiscardUnknown(m)
}

var xxx_messageInfo_ResponseStatus_Error proto.InternalMessageInfo

func (m *ResponseStatus_Error) GetSlug() string {
	if m != nil {
		return m.Slug
	}
	return ""
}

func (m *ResponseStatus_Error) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

type RequestSession struct {
	AccessToken string `protobuf:"bytes,1,opt,name=access_token,json=accessToken,proto3" json:"access_token,omitempty"`
	Locale      string `protobuf:"bytes,2,opt,name=locale,proto3" json:"locale,omitempty"`
	// Validate if it's not in the future
	Timestamp            int64    `protobuf:"varint,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RequestSession) Reset()         { *m = RequestSession{} }
func (m *RequestSession) String() string { return proto.CompactTextString(m) }
func (*RequestSession) ProtoMessage()    {}
func (*RequestSession) Descriptor() ([]byte, []int) {
	return fileDescriptor_cc8776310bb7bdb9, []int{1}
}

func (m *RequestSession) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RequestSession.Unmarshal(m, b)
}
func (m *RequestSession) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RequestSession.Marshal(b, m, deterministic)
}
func (m *RequestSession) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RequestSession.Merge(m, src)
}
func (m *RequestSession) XXX_Size() int {
	return xxx_messageInfo_RequestSession.Size(m)
}
func (m *RequestSession) XXX_DiscardUnknown() {
	xxx_messageInfo_RequestSession.DiscardUnknown(m)
}

var xxx_messageInfo_RequestSession proto.InternalMessageInfo

func (m *RequestSession) GetAccessToken() string {
	if m != nil {
		return m.AccessToken
	}
	return ""
}

func (m *RequestSession) GetLocale() string {
	if m != nil {
		return m.Locale
	}
	return ""
}

func (m *RequestSession) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func init() {
	proto.RegisterType((*ResponseStatus)(nil), "sso.common.ResponseStatus")
	proto.RegisterType((*ResponseStatus_Error)(nil), "sso.common.ResponseStatus.Error")
	proto.RegisterType((*RequestSession)(nil), "sso.common.RequestSession")
}

func init() { proto.RegisterFile("sso-common/sso-common.proto", fileDescriptor_cc8776310bb7bdb9) }

var fileDescriptor_cc8776310bb7bdb9 = []byte{
	// 244 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0xc1, 0x4a, 0xc3, 0x40,
	0x14, 0x45, 0x89, 0xd5, 0x48, 0x5f, 0xc5, 0xc5, 0x2c, 0x24, 0xa8, 0x8b, 0xd8, 0x55, 0x36, 0x9d,
	0x88, 0x22, 0xb8, 0x16, 0xfc, 0x81, 0xd4, 0x95, 0x1b, 0x99, 0xc6, 0x47, 0x1c, 0x9a, 0x99, 0x37,
	0xe6, 0x4e, 0x3e, 0xc0, 0x3f, 0x97, 0xa4, 0x29, 0xd1, 0xdd, 0xbb, 0x87, 0xc3, 0x65, 0xe6, 0xd2,
	0x0d, 0x20, 0x9b, 0x5a, 0x9c, 0x13, 0x5f, 0xce, 0xa7, 0x0e, 0x9d, 0x44, 0x51, 0x04, 0x88, 0x3e,
	0x90, 0xf5, 0x4f, 0x42, 0x97, 0x15, 0x23, 0x88, 0x07, 0x6f, 0xa3, 0x89, 0x3d, 0xd4, 0x33, 0xa5,
	0xdc, 0x75, 0xd2, 0x21, 0x4b, 0xf2, 0x45, 0xb1, 0x7a, 0xc8, 0xf5, 0xec, 0xeb, 0xff, 0xae, 0x7e,
	0x1d, 0xc4, 0x6a, 0xf2, 0xaf, 0x9f, 0xe8, 0x6c, 0x04, 0x4a, 0xd1, 0x29, 0xda, 0xbe, 0xc9, 0x92,
	0x3c, 0x29, 0x96, 0xd5, 0x78, 0xab, 0x8c, 0xce, 0x1d, 0x03, 0xa6, 0xe1, 0xec, 0x64, 0xc4, 0xc7,
	0xb8, 0xb6, 0xc3, 0x13, 0xbe, 0x7b, 0x46, 0xdc, 0x32, 0x60, 0xc5, 0xab, 0x3b, 0xba, 0x30, 0x75,
	0xcd, 0xc0, 0x47, 0x94, 0x3d, 0xfb, 0xa9, 0x67, 0x75, 0x60, 0x6f, 0x03, 0x52, 0x57, 0x94, 0xb6,
	0x52, 0x9b, 0xf6, 0xd8, 0x36, 0x25, 0x75, 0x4b, 0xcb, 0x68, 0x1d, 0x23, 0x1a, 0x17, 0xb2, 0x45,
	0x9e, 0x14, 0x8b, 0x6a, 0x06, 0x2f, 0xf7, 0xef, 0xba, 0xb1, 0xf1, 0xab, 0xdf, 0x0d, 0xff, 0x29,
	0x1b, 0xe3, 0x8c, 0x04, 0x94, 0x4e, 0xbc, 0x6c, 0x00, 0x29, 0xc3, 0xbe, 0x29, 0xed, 0x67, 0xfb,
	0x67, 0xb2, 0x5d, 0x3a, 0x6e, 0xf6, 0xf8, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x49, 0xf9, 0xc9, 0x07,
	0x52, 0x01, 0x00, 0x00,
}
