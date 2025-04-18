// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: helloworld/v1/student.proto

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
	Student_AddStudent_FullMethodName   = "/helloworld.v1.Student/AddStudent"
	Student_GetStudent_FullMethodName   = "/helloworld.v1.Student/GetStudent"
	Student_ListStudents_FullMethodName = "/helloworld.v1.Student/ListStudents"
	Student_MessageSend_FullMethodName  = "/helloworld.v1.Student/MessageSend"
	Student_ActivityGet_FullMethodName  = "/helloworld.v1.Student/ActivityGet"
	Student_TimeGet_FullMethodName      = "/helloworld.v1.Student/TimeGet"
	Student_Invitation_FullMethodName   = "/helloworld.v1.Student/Invitation"
)

// StudentClient is the client API for Student service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type StudentClient interface {
	AddStudent(ctx context.Context, in *AddStudentRequest, opts ...grpc.CallOption) (*AddStudentResponse, error)
	GetStudent(ctx context.Context, in *GetStudentRequest, opts ...grpc.CallOption) (*GetStudentRespose, error)
	ListStudents(ctx context.Context, in *ListStudentsRequest, opts ...grpc.CallOption) (*ListStudentsResponse, error)
	MessageSend(ctx context.Context, in *InvitationRequest, opts ...grpc.CallOption) (*InvitationResponse, error)
	ActivityGet(ctx context.Context, in *InvitationRequest, opts ...grpc.CallOption) (*InvitationResponse, error)
	TimeGet(ctx context.Context, in *InvitationRequest, opts ...grpc.CallOption) (*InvitationResponse, error)
	Invitation(ctx context.Context, in *InvitationRequest, opts ...grpc.CallOption) (*InvitationResponse, error)
}

type studentClient struct {
	cc grpc.ClientConnInterface
}

func NewStudentClient(cc grpc.ClientConnInterface) StudentClient {
	return &studentClient{cc}
}

func (c *studentClient) AddStudent(ctx context.Context, in *AddStudentRequest, opts ...grpc.CallOption) (*AddStudentResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AddStudentResponse)
	err := c.cc.Invoke(ctx, Student_AddStudent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studentClient) GetStudent(ctx context.Context, in *GetStudentRequest, opts ...grpc.CallOption) (*GetStudentRespose, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetStudentRespose)
	err := c.cc.Invoke(ctx, Student_GetStudent_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studentClient) ListStudents(ctx context.Context, in *ListStudentsRequest, opts ...grpc.CallOption) (*ListStudentsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(ListStudentsResponse)
	err := c.cc.Invoke(ctx, Student_ListStudents_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studentClient) MessageSend(ctx context.Context, in *InvitationRequest, opts ...grpc.CallOption) (*InvitationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InvitationResponse)
	err := c.cc.Invoke(ctx, Student_MessageSend_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studentClient) ActivityGet(ctx context.Context, in *InvitationRequest, opts ...grpc.CallOption) (*InvitationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InvitationResponse)
	err := c.cc.Invoke(ctx, Student_ActivityGet_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studentClient) TimeGet(ctx context.Context, in *InvitationRequest, opts ...grpc.CallOption) (*InvitationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InvitationResponse)
	err := c.cc.Invoke(ctx, Student_TimeGet_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *studentClient) Invitation(ctx context.Context, in *InvitationRequest, opts ...grpc.CallOption) (*InvitationResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(InvitationResponse)
	err := c.cc.Invoke(ctx, Student_Invitation_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StudentServer is the server API for Student service.
// All implementations must embed UnimplementedStudentServer
// for forward compatibility.
type StudentServer interface {
	AddStudent(context.Context, *AddStudentRequest) (*AddStudentResponse, error)
	GetStudent(context.Context, *GetStudentRequest) (*GetStudentRespose, error)
	ListStudents(context.Context, *ListStudentsRequest) (*ListStudentsResponse, error)
	MessageSend(context.Context, *InvitationRequest) (*InvitationResponse, error)
	ActivityGet(context.Context, *InvitationRequest) (*InvitationResponse, error)
	TimeGet(context.Context, *InvitationRequest) (*InvitationResponse, error)
	Invitation(context.Context, *InvitationRequest) (*InvitationResponse, error)
	mustEmbedUnimplementedStudentServer()
}

// UnimplementedStudentServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedStudentServer struct{}

func (UnimplementedStudentServer) AddStudent(context.Context, *AddStudentRequest) (*AddStudentResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddStudent not implemented")
}
func (UnimplementedStudentServer) GetStudent(context.Context, *GetStudentRequest) (*GetStudentRespose, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStudent not implemented")
}
func (UnimplementedStudentServer) ListStudents(context.Context, *ListStudentsRequest) (*ListStudentsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListStudents not implemented")
}
func (UnimplementedStudentServer) MessageSend(context.Context, *InvitationRequest) (*InvitationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MessageSend not implemented")
}
func (UnimplementedStudentServer) ActivityGet(context.Context, *InvitationRequest) (*InvitationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ActivityGet not implemented")
}
func (UnimplementedStudentServer) TimeGet(context.Context, *InvitationRequest) (*InvitationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TimeGet not implemented")
}
func (UnimplementedStudentServer) Invitation(context.Context, *InvitationRequest) (*InvitationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Invitation not implemented")
}
func (UnimplementedStudentServer) mustEmbedUnimplementedStudentServer() {}
func (UnimplementedStudentServer) testEmbeddedByValue()                 {}

// UnsafeStudentServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to StudentServer will
// result in compilation errors.
type UnsafeStudentServer interface {
	mustEmbedUnimplementedStudentServer()
}

func RegisterStudentServer(s grpc.ServiceRegistrar, srv StudentServer) {
	// If the following call pancis, it indicates UnimplementedStudentServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Student_ServiceDesc, srv)
}

func _Student_AddStudent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddStudentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServer).AddStudent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Student_AddStudent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServer).AddStudent(ctx, req.(*AddStudentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Student_GetStudent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetStudentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServer).GetStudent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Student_GetStudent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServer).GetStudent(ctx, req.(*GetStudentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Student_ListStudents_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListStudentsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServer).ListStudents(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Student_ListStudents_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServer).ListStudents(ctx, req.(*ListStudentsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Student_MessageSend_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InvitationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServer).MessageSend(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Student_MessageSend_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServer).MessageSend(ctx, req.(*InvitationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Student_ActivityGet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InvitationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServer).ActivityGet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Student_ActivityGet_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServer).ActivityGet(ctx, req.(*InvitationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Student_TimeGet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InvitationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServer).TimeGet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Student_TimeGet_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServer).TimeGet(ctx, req.(*InvitationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Student_Invitation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InvitationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StudentServer).Invitation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Student_Invitation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StudentServer).Invitation(ctx, req.(*InvitationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Student_ServiceDesc is the grpc.ServiceDesc for Student service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Student_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "helloworld.v1.Student",
	HandlerType: (*StudentServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddStudent",
			Handler:    _Student_AddStudent_Handler,
		},
		{
			MethodName: "GetStudent",
			Handler:    _Student_GetStudent_Handler,
		},
		{
			MethodName: "ListStudents",
			Handler:    _Student_ListStudents_Handler,
		},
		{
			MethodName: "MessageSend",
			Handler:    _Student_MessageSend_Handler,
		},
		{
			MethodName: "ActivityGet",
			Handler:    _Student_ActivityGet_Handler,
		},
		{
			MethodName: "TimeGet",
			Handler:    _Student_TimeGet_Handler,
		},
		{
			MethodName: "Invitation",
			Handler:    _Student_Invitation_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "helloworld/v1/student.proto",
}
