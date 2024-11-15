// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v3.12.4
// source: grpc/pb.proto

package proto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	RicardService_Request_FullMethodName = "/RicardService/Request"
)

// RicardServiceClient is the client API for RicardService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RicardServiceClient interface {
	Request(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Empty, error)
}

type ricardServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRicardServiceClient(cc grpc.ClientConnInterface) RicardServiceClient {
	return &ricardServiceClient{cc}
}

func (c *ricardServiceClient) Request(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Empty, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Empty)
	err := c.cc.Invoke(ctx, RicardService_Request_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RicardServiceServer is the server API for RicardService service.
// All implementations must embed UnimplementedRicardServiceServer
// for forward compatibility.
type RicardServiceServer interface {
	Request(context.Context, *Message) (*Empty, error)
	mustEmbedUnimplementedRicardServiceServer()
}

// UnimplementedRicardServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedRicardServiceServer struct{}

func (UnimplementedRicardServiceServer) Request(context.Context, *Message) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Request not implemented")
}
func (UnimplementedRicardServiceServer) mustEmbedUnimplementedRicardServiceServer() {}
func (UnimplementedRicardServiceServer) testEmbeddedByValue()                       {}

// UnsafeRicardServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RicardServiceServer will
// result in compilation errors.
type UnsafeRicardServiceServer interface {
	mustEmbedUnimplementedRicardServiceServer()
}

func RegisterRicardServiceServer(s grpc.ServiceRegistrar, srv RicardServiceServer) {
	// If the following call pancis, it indicates UnimplementedRicardServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&RicardService_ServiceDesc, srv)
}

func _RicardService_Request_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RicardServiceServer).Request(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: RicardService_Request_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RicardServiceServer).Request(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

// RicardService_ServiceDesc is the grpc.ServiceDesc for RicardService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RicardService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "RicardService",
	HandlerType: (*RicardServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Request",
			Handler:    _RicardService_Request_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/pb.proto",
}
