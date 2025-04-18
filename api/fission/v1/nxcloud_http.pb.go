// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.8.3
// - protoc             v5.29.3
// source: fission/v1/nxcloud.proto

package v1

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationNXCloudUserAttendInfo = "/fission.v1.NXCloud/UserAttendInfo"
const OperationNXCloudUserAttendInfo2 = "/fission.v1.NXCloud/UserAttendInfo2"

type NXCloudHTTPServer interface {
	UserAttendInfo(context.Context, *UserAttendInfoRequest) (*UserAttendInfoResponse, error)
	// UserAttendInfo2 UserAttendInfo2 implements v1.NXCloudHTTPServer.
	// webhook
	UserAttendInfo2(context.Context, *UserAttendInfoRequest) (*UserAttendInfoResponse, error)
}

func RegisterNXCloudHTTPServer(s *http.Server, srv NXCloudHTTPServer) {
	r := s.Route("/")
	r.POST("/events/mlbb25031gateway/activity/userAttendInfo", _NXCloud_UserAttendInfo0_HTTP_Handler(srv))
	r.POST("/events/mlbb25031gateway/activity/userAttendInfo2", _NXCloud_UserAttendInfo20_HTTP_Handler(srv))
}

func _NXCloud_UserAttendInfo0_HTTP_Handler(srv NXCloudHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UserAttendInfoRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationNXCloudUserAttendInfo)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.UserAttendInfo(ctx, req.(*UserAttendInfoRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*UserAttendInfoResponse)
		return ctx.Result(200, reply)
	}
}

func _NXCloud_UserAttendInfo20_HTTP_Handler(srv NXCloudHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in UserAttendInfoRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationNXCloudUserAttendInfo2)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.UserAttendInfo2(ctx, req.(*UserAttendInfoRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*UserAttendInfoResponse)
		return ctx.Result(200, reply)
	}
}

type NXCloudHTTPClient interface {
	UserAttendInfo(ctx context.Context, req *UserAttendInfoRequest, opts ...http.CallOption) (rsp *UserAttendInfoResponse, err error)
	UserAttendInfo2(ctx context.Context, req *UserAttendInfoRequest, opts ...http.CallOption) (rsp *UserAttendInfoResponse, err error)
}

type NXCloudHTTPClientImpl struct {
	cc *http.Client
}

func NewNXCloudHTTPClient(client *http.Client) NXCloudHTTPClient {
	return &NXCloudHTTPClientImpl{client}
}

func (c *NXCloudHTTPClientImpl) UserAttendInfo(ctx context.Context, in *UserAttendInfoRequest, opts ...http.CallOption) (*UserAttendInfoResponse, error) {
	var out UserAttendInfoResponse
	pattern := "/events/mlbb25031gateway/activity/userAttendInfo"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationNXCloudUserAttendInfo))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *NXCloudHTTPClientImpl) UserAttendInfo2(ctx context.Context, in *UserAttendInfoRequest, opts ...http.CallOption) (*UserAttendInfoResponse, error) {
	var out UserAttendInfoResponse
	pattern := "/events/mlbb25031gateway/activity/userAttendInfo2"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationNXCloudUserAttendInfo2))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
