// Code generated by protoc-gen-go. DO NOT EDIT.
// source: chordio.proto

package pb

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

//message Node {
//    uint32 id = 1;
//    string addr = 2;
//    uint32 port = 3;
//    Node prev = 4;
//    Node next = 5;
//}
//
//message FingerTable {
//    repeated FingerTableEntry entries = 1;
//}
//
//message FingerTableEntry {
//    uint32 start = 1;
//    uint32 end = 2;
//    Node successor = 3;
//}
//
type GetIDRequest struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetIDRequest) Reset()         { *m = GetIDRequest{} }
func (m *GetIDRequest) String() string { return proto.CompactTextString(m) }
func (*GetIDRequest) ProtoMessage()    {}
func (*GetIDRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d572f8aeec35237e, []int{0}
}

func (m *GetIDRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetIDRequest.Unmarshal(m, b)
}
func (m *GetIDRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetIDRequest.Marshal(b, m, deterministic)
}
func (m *GetIDRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetIDRequest.Merge(m, src)
}
func (m *GetIDRequest) XXX_Size() int {
	return xxx_messageInfo_GetIDRequest.Size(m)
}
func (m *GetIDRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetIDRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetIDRequest proto.InternalMessageInfo

type GetIDResponse struct {
	Id                   uint64   `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetIDResponse) Reset()         { *m = GetIDResponse{} }
func (m *GetIDResponse) String() string { return proto.CompactTextString(m) }
func (*GetIDResponse) ProtoMessage()    {}
func (*GetIDResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d572f8aeec35237e, []int{1}
}

func (m *GetIDResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetIDResponse.Unmarshal(m, b)
}
func (m *GetIDResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetIDResponse.Marshal(b, m, deterministic)
}
func (m *GetIDResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetIDResponse.Merge(m, src)
}
func (m *GetIDResponse) XXX_Size() int {
	return xxx_messageInfo_GetIDResponse.Size(m)
}
func (m *GetIDResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetIDResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetIDResponse proto.InternalMessageInfo

func (m *GetIDResponse) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func init() {
	proto.RegisterType((*GetIDRequest)(nil), "GetIDRequest")
	proto.RegisterType((*GetIDResponse)(nil), "GetIDResponse")
}

func init() {
	proto.RegisterFile("chordio.proto", fileDescriptor_d572f8aeec35237e)
}

var fileDescriptor_d572f8aeec35237e = []byte{
	// 118 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x4d, 0xce, 0xc8, 0x2f,
	0x4a, 0xc9, 0xcc, 0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x57, 0xe2, 0xe3, 0xe2, 0x71, 0x4f, 0x2d,
	0xf1, 0x74, 0x09, 0x4a, 0x2d, 0x2c, 0x4d, 0x2d, 0x2e, 0x51, 0x92, 0xe7, 0xe2, 0x85, 0xf2, 0x8b,
	0x0b, 0xf2, 0xf3, 0x8a, 0x53, 0x85, 0xf8, 0xb8, 0x98, 0x32, 0x53, 0x24, 0x18, 0x15, 0x18, 0x35,
	0x58, 0x82, 0x80, 0x2c, 0x23, 0x43, 0x2e, 0x56, 0x67, 0x90, 0x09, 0x42, 0x1a, 0x5c, 0xac, 0x60,
	0x95, 0x42, 0xbc, 0x7a, 0xc8, 0x26, 0x48, 0xf1, 0xe9, 0xa1, 0x18, 0xa0, 0xc4, 0xe0, 0xc4, 0x12,
	0xc5, 0x54, 0x90, 0x94, 0xc4, 0x06, 0xb6, 0xd0, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0xc4, 0x76,
	0x6a, 0x14, 0x81, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// ChordClient is the client API for Chord service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type ChordClient interface {
	GetID(ctx context.Context, in *GetIDRequest, opts ...grpc.CallOption) (*GetIDResponse, error)
}

type chordClient struct {
	cc grpc.ClientConnInterface
}

func NewChordClient(cc grpc.ClientConnInterface) ChordClient {
	return &chordClient{cc}
}

func (c *chordClient) GetID(ctx context.Context, in *GetIDRequest, opts ...grpc.CallOption) (*GetIDResponse, error) {
	out := new(GetIDResponse)
	err := c.cc.Invoke(ctx, "/Chord/GetID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChordServer is the server API for Chord service.
type ChordServer interface {
	GetID(context.Context, *GetIDRequest) (*GetIDResponse, error)
}

// UnimplementedChordServer can be embedded to have forward compatible implementations.
type UnimplementedChordServer struct {
}

func (*UnimplementedChordServer) GetID(ctx context.Context, req *GetIDRequest) (*GetIDResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetID not implemented")
}

func RegisterChordServer(s *grpc.Server, srv ChordServer) {
	s.RegisterService(&_Chord_serviceDesc, srv)
}

func _Chord_GetID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChordServer).GetID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Chord/GetID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChordServer).GetID(ctx, req.(*GetIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Chord_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Chord",
	HandlerType: (*ChordServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetID",
			Handler:    _Chord_GetID_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "chordio.proto",
}
