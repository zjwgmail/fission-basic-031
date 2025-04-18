// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: fission/v1/cdk.proto

package v1

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
	CDK_GetCDK_FullMethodName    = "/fission.v1.CDK/GetCDK"
	CDK_ImportCDK_FullMethodName = "/fission.v1.CDK/ImportCDK"
	CDK_CDKTest_FullMethodName   = "/fission.v1.CDK/CDKTest"
)

// CDKClient is the client API for CDK service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CDKClient interface {
	GetCDK(ctx context.Context, in *GetCDKRequest, opts ...grpc.CallOption) (*GetCDKResponse, error)
	ImportCDK(ctx context.Context, in *ImportCDKRequest, opts ...grpc.CallOption) (*ImportCDKResponse, error)
	CDKTest(ctx context.Context, in *CDKTestRequest, opts ...grpc.CallOption) (*CDKTestResponse, error)
}

type cDKClient struct {
	cc grpc.ClientConnInterface
}

func NewCDKClient(cc grpc.ClientConnInterface) CDKClient {
	return &cDKClient{cc}
}

func (c *cDKClient) GetCDK(ctx context.Context, in *GetCDKRequest, opts ...grpc.CallOption) (*GetCDKResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetCDKResponse)
	err := c.cc.Invoke(ctx, CDK_GetCDK_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cDKClient) ImportCDK(ctx context.Context, in *ImportCDKRequest, opts ...grpc.CallOption) (*ImportCDKResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ImportCDKResponse)
	err := c.cc.Invoke(ctx, CDK_ImportCDK_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cDKClient) CDKTest(ctx context.Context, in *CDKTestRequest, opts ...grpc.CallOption) (*CDKTestResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CDKTestResponse)
	err := c.cc.Invoke(ctx, CDK_CDKTest_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CDKServer is the server API for CDK service.
// All implementations must embed UnimplementedCDKServer
// for forward compatibility.
type CDKServer interface {
	GetCDK(context.Context, *GetCDKRequest) (*GetCDKResponse, error)
	ImportCDK(context.Context, *ImportCDKRequest) (*ImportCDKResponse, error)
	CDKTest(context.Context, *CDKTestRequest) (*CDKTestResponse, error)
	mustEmbedUnimplementedCDKServer()
}

// UnimplementedCDKServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedCDKServer struct{}

func (UnimplementedCDKServer) GetCDK(context.Context, *GetCDKRequest) (*GetCDKResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCDK not implemented")
}
func (UnimplementedCDKServer) ImportCDK(context.Context, *ImportCDKRequest) (*ImportCDKResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ImportCDK not implemented")
}
func (UnimplementedCDKServer) CDKTest(context.Context, *CDKTestRequest) (*CDKTestResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CDKTest not implemented")
}
func (UnimplementedCDKServer) mustEmbedUnimplementedCDKServer() {}
func (UnimplementedCDKServer) testEmbeddedByValue()             {}

// UnsafeCDKServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CDKServer will
// result in compilation errors.
type UnsafeCDKServer interface {
	mustEmbedUnimplementedCDKServer()
}

func RegisterCDKServer(s grpc.ServiceRegistrar, srv CDKServer) {
	// If the following call pancis, it indicates UnimplementedCDKServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&CDK_ServiceDesc, srv)
}

func _CDK_GetCDK_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCDKRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CDKServer).GetCDK(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CDK_GetCDK_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CDKServer).GetCDK(ctx, req.(*GetCDKRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CDK_ImportCDK_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ImportCDKRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CDKServer).ImportCDK(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CDK_ImportCDK_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CDKServer).ImportCDK(ctx, req.(*ImportCDKRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CDK_CDKTest_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CDKTestRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CDKServer).CDKTest(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CDK_CDKTest_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CDKServer).CDKTest(ctx, req.(*CDKTestRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CDK_ServiceDesc is the grpc.ServiceDesc for CDK service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CDK_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "fission.v1.CDK",
	HandlerType: (*CDKServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetCDK",
			Handler:    _CDK_GetCDK_Handler,
		},
		{
			MethodName: "ImportCDK",
			Handler:    _CDK_ImportCDK_Handler,
		},
		{
			MethodName: "CDKTest",
			Handler:    _CDK_CDKTest_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fission/v1/cdk.proto",
}
