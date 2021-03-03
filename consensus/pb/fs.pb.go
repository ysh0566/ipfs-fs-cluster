// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: consensus/pb/fs.proto

package pb

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
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
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

type Instruction_Code int32

const (
	Instruction_CP    Instruction_Code = 0
	Instruction_MV    Instruction_Code = 1
	Instruction_RM    Instruction_Code = 2
	Instruction_MKDIR Instruction_Code = 3
	Instruction_Ls    Instruction_Code = 4
)

var Instruction_Code_name = map[int32]string{
	0: "CP",
	1: "MV",
	2: "RM",
	3: "MKDIR",
	4: "Ls",
}

var Instruction_Code_value = map[string]int32{
	"CP":    0,
	"MV":    1,
	"RM":    2,
	"MKDIR": 3,
	"Ls":    4,
}

func (x Instruction_Code) String() string {
	return proto.EnumName(Instruction_Code_name, int32(x))
}

func (Instruction_Code) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_0e1a8c64c0f1b0bd, []int{2, 0}
}

type Ctx struct {
	Pre                  string   `protobuf:"bytes,1,opt,name=pre,proto3" json:"pre,omitempty"`
	Next                 string   `protobuf:"bytes,2,opt,name=next,proto3" json:"next,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Ctx) Reset()         { *m = Ctx{} }
func (m *Ctx) String() string { return proto.CompactTextString(m) }
func (*Ctx) ProtoMessage()    {}
func (*Ctx) Descriptor() ([]byte, []int) {
	return fileDescriptor_0e1a8c64c0f1b0bd, []int{0}
}
func (m *Ctx) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Ctx.Unmarshal(m, b)
}
func (m *Ctx) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Ctx.Marshal(b, m, deterministic)
}
func (m *Ctx) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Ctx.Merge(m, src)
}
func (m *Ctx) XXX_Size() int {
	return xxx_messageInfo_Ctx.Size(m)
}
func (m *Ctx) XXX_DiscardUnknown() {
	xxx_messageInfo_Ctx.DiscardUnknown(m)
}

var xxx_messageInfo_Ctx proto.InternalMessageInfo

func (m *Ctx) GetPre() string {
	if m != nil {
		return m.Pre
	}
	return ""
}

func (m *Ctx) GetNext() string {
	if m != nil {
		return m.Next
	}
	return ""
}

type Empty struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Empty) Reset()         { *m = Empty{} }
func (m *Empty) String() string { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()    {}
func (*Empty) Descriptor() ([]byte, []int) {
	return fileDescriptor_0e1a8c64c0f1b0bd, []int{1}
}
func (m *Empty) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Empty.Unmarshal(m, b)
}
func (m *Empty) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Empty.Marshal(b, m, deterministic)
}
func (m *Empty) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Empty.Merge(m, src)
}
func (m *Empty) XXX_Size() int {
	return xxx_messageInfo_Empty.Size(m)
}
func (m *Empty) XXX_DiscardUnknown() {
	xxx_messageInfo_Empty.DiscardUnknown(m)
}

var xxx_messageInfo_Empty proto.InternalMessageInfo

type Instruction struct {
	Code                 Instruction_Code `protobuf:"varint,1,opt,name=code,proto3,enum=pb.Instruction_Code" json:"code,omitempty"`
	Params               []string         `protobuf:"bytes,2,rep,name=params,proto3" json:"params,omitempty"`
	Node                 []byte           `protobuf:"bytes,3,opt,name=node,proto3" json:"node,omitempty"`
	XXX_NoUnkeyedLiteral struct{}         `json:"-"`
	XXX_unrecognized     []byte           `json:"-"`
	XXX_sizecache        int32            `json:"-"`
}

func (m *Instruction) Reset()         { *m = Instruction{} }
func (m *Instruction) String() string { return proto.CompactTextString(m) }
func (*Instruction) ProtoMessage()    {}
func (*Instruction) Descriptor() ([]byte, []int) {
	return fileDescriptor_0e1a8c64c0f1b0bd, []int{2}
}
func (m *Instruction) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Instruction.Unmarshal(m, b)
}
func (m *Instruction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Instruction.Marshal(b, m, deterministic)
}
func (m *Instruction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Instruction.Merge(m, src)
}
func (m *Instruction) XXX_Size() int {
	return xxx_messageInfo_Instruction.Size(m)
}
func (m *Instruction) XXX_DiscardUnknown() {
	xxx_messageInfo_Instruction.DiscardUnknown(m)
}

var xxx_messageInfo_Instruction proto.InternalMessageInfo

func (m *Instruction) GetCode() Instruction_Code {
	if m != nil {
		return m.Code
	}
	return Instruction_CP
}

func (m *Instruction) GetParams() []string {
	if m != nil {
		return m.Params
	}
	return nil
}

func (m *Instruction) GetNode() []byte {
	if m != nil {
		return m.Node
	}
	return nil
}

type Instructions struct {
	Instruction          []*Instruction `protobuf:"bytes,1,rep,name=instruction,proto3" json:"instruction,omitempty"`
	Ctx                  *Ctx           `protobuf:"bytes,2,opt,name=ctx,proto3" json:"ctx,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *Instructions) Reset()         { *m = Instructions{} }
func (m *Instructions) String() string { return proto.CompactTextString(m) }
func (*Instructions) ProtoMessage()    {}
func (*Instructions) Descriptor() ([]byte, []int) {
	return fileDescriptor_0e1a8c64c0f1b0bd, []int{3}
}
func (m *Instructions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Instructions.Unmarshal(m, b)
}
func (m *Instructions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Instructions.Marshal(b, m, deterministic)
}
func (m *Instructions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Instructions.Merge(m, src)
}
func (m *Instructions) XXX_Size() int {
	return xxx_messageInfo_Instructions.Size(m)
}
func (m *Instructions) XXX_DiscardUnknown() {
	xxx_messageInfo_Instructions.DiscardUnknown(m)
}

var xxx_messageInfo_Instructions proto.InternalMessageInfo

func (m *Instructions) GetInstruction() []*Instruction {
	if m != nil {
		return m.Instruction
	}
	return nil
}

func (m *Instructions) GetCtx() *Ctx {
	if m != nil {
		return m.Ctx
	}
	return nil
}

func init() {
	proto.RegisterEnum("pb.Instruction_Code", Instruction_Code_name, Instruction_Code_value)
	proto.RegisterType((*Ctx)(nil), "pb.Ctx")
	proto.RegisterType((*Empty)(nil), "pb.Empty")
	proto.RegisterType((*Instruction)(nil), "pb.Instruction")
	proto.RegisterType((*Instructions)(nil), "pb.Instructions")
}

func init() { proto.RegisterFile("consensus/pb/fs.proto", fileDescriptor_0e1a8c64c0f1b0bd) }

var fileDescriptor_0e1a8c64c0f1b0bd = []byte{
	// 288 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0x41, 0x4b, 0xc3, 0x30,
	0x14, 0xc7, 0x97, 0xa4, 0x5b, 0xe9, 0xeb, 0xd4, 0x10, 0x54, 0xaa, 0xa7, 0xd2, 0x53, 0x40, 0xe8,
	0x58, 0x3d, 0x79, 0xb5, 0xee, 0x30, 0xb4, 0x20, 0x39, 0x78, 0x10, 0x2f, 0x6b, 0x17, 0x61, 0x87,
	0x36, 0xa1, 0x49, 0xa1, 0x7e, 0x11, 0x3f, 0xaf, 0x24, 0x53, 0x2c, 0x3b, 0xbd, 0xff, 0xfb, 0xe5,
	0xff, 0xde, 0xcb, 0x7b, 0x70, 0xd5, 0xa8, 0xce, 0xc8, 0xce, 0x0c, 0x66, 0xa5, 0xeb, 0xd5, 0xa7,
	0xc9, 0x75, 0xaf, 0xac, 0x62, 0x58, 0xd7, 0xd9, 0x1d, 0x90, 0xd2, 0x8e, 0x8c, 0x02, 0xd1, 0xbd,
	0x4c, 0x50, 0x8a, 0x78, 0x24, 0x9c, 0x64, 0x0c, 0x82, 0x4e, 0x8e, 0x36, 0xc1, 0x1e, 0x79, 0x9d,
	0x85, 0x30, 0xdf, 0xb4, 0xda, 0x7e, 0x65, 0xdf, 0x08, 0xe2, 0x6d, 0x67, 0x6c, 0x3f, 0x34, 0xf6,
	0xa0, 0x3a, 0xc6, 0x21, 0x68, 0xd4, 0xfe, 0x58, 0x7f, 0x5e, 0x5c, 0xe6, 0xba, 0xce, 0x27, 0xcf,
	0x79, 0xa9, 0xf6, 0x52, 0x78, 0x07, 0xbb, 0x86, 0x85, 0xde, 0xf5, 0xbb, 0xd6, 0x24, 0x38, 0x25,
	0x3c, 0x12, 0xbf, 0x99, 0x1f, 0xe7, 0x3a, 0x90, 0x14, 0xf1, 0xa5, 0xf0, 0x3a, 0x5b, 0x43, 0xe0,
	0x2a, 0xd9, 0x02, 0x70, 0xf9, 0x4a, 0x67, 0x2e, 0x56, 0x6f, 0x14, 0xb9, 0x28, 0x2a, 0x8a, 0x59,
	0x04, 0xf3, 0xea, 0xf9, 0x69, 0x2b, 0x28, 0x71, 0xe8, 0xc5, 0xd0, 0x20, 0xfb, 0x80, 0xe5, 0x64,
	0xb0, 0x61, 0x6b, 0x88, 0x0f, 0xff, 0x79, 0x82, 0x52, 0xc2, 0xe3, 0xe2, 0xe2, 0xe4, 0x7f, 0x62,
	0xea, 0x61, 0x37, 0x40, 0x1a, 0x3b, 0xfa, 0xbd, 0xe3, 0x22, 0x74, 0xd6, 0xd2, 0x8e, 0xc2, 0xb1,
	0xe2, 0x01, 0xce, 0x84, 0x6c, 0x95, 0x95, 0x9b, 0x51, 0x36, 0x83, 0x95, 0x8c, 0x43, 0xf8, 0x27,
	0xe9, 0x49, 0x53, 0x73, 0x1b, 0x39, 0x72, 0xbc, 0xd7, 0xec, 0x11, 0xbf, 0xcf, 0xea, 0x85, 0x3f,
	0xfb, 0xfd, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x3d, 0x73, 0x01, 0xa7, 0x8f, 0x01, 0x00, 0x00,
}
