// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.8.3
// - protoc             v5.29.3
// source: fission/v1/help_code.proto

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

const OperationHelpCodeGetActivityInfo = "/helloworld.v1.HelpCode/GetActivityInfo"
const OperationHelpCodeHCTest = "/helloworld.v1.HelpCode/HCTest"
const OperationHelpCodePreheatHelpCode = "/helloworld.v1.HelpCode/PreheatHelpCode"
const OperationHelpCodeRepairHelpCode = "/helloworld.v1.HelpCode/RepairHelpCode"

type HelpCodeHTTPServer interface {
	GetActivityInfo(context.Context, *GetActivityInfoRequest) (*GetActivityInfoResponse, error)
	HCTest(context.Context, *HCTestRequest) (*HCTestResponse, error)
	PreheatHelpCode(context.Context, *PreheatHelpCodeRequest) (*PreheatHelpCodeResponse, error)
	RepairHelpCode(context.Context, *RepairHelpCodeRequest) (*RepairHelpCodeResponse, error)
}

func RegisterHelpCodeHTTPServer(s *http.Server, srv HelpCodeHTTPServer) {
	r := s.Route("/")
	r.POST("/events/mlbb25031job/activity/helpCode/preHeat", _HelpCode_PreheatHelpCode0_HTTP_Handler(srv))
	r.POST("/events/mlbb25031job/activity/helpCode/repair", _HelpCode_RepairHelpCode0_HTTP_Handler(srv))
	r.GET("/events/mlbb25031job/activity/helpCode/test", _HelpCode_HCTest0_HTTP_Handler(srv))
	r.GET("/events/mlbb25031job/activity/getactivityinfo", _HelpCode_GetActivityInfo0_HTTP_Handler(srv))
}

func _HelpCode_PreheatHelpCode0_HTTP_Handler(srv HelpCodeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in PreheatHelpCodeRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationHelpCodePreheatHelpCode)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.PreheatHelpCode(ctx, req.(*PreheatHelpCodeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*PreheatHelpCodeResponse)
		return ctx.Result(200, reply)
	}
}

func _HelpCode_RepairHelpCode0_HTTP_Handler(srv HelpCodeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in RepairHelpCodeRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationHelpCodeRepairHelpCode)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.RepairHelpCode(ctx, req.(*RepairHelpCodeRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*RepairHelpCodeResponse)
		return ctx.Result(200, reply)
	}
}

func _HelpCode_HCTest0_HTTP_Handler(srv HelpCodeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in HCTestRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationHelpCodeHCTest)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.HCTest(ctx, req.(*HCTestRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*HCTestResponse)
		return ctx.Result(200, reply)
	}
}

func _HelpCode_GetActivityInfo0_HTTP_Handler(srv HelpCodeHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetActivityInfoRequest
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationHelpCodeGetActivityInfo)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetActivityInfo(ctx, req.(*GetActivityInfoRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetActivityInfoResponse)
		return ctx.Result(200, reply)
	}
}

type HelpCodeHTTPClient interface {
	GetActivityInfo(ctx context.Context, req *GetActivityInfoRequest, opts ...http.CallOption) (rsp *GetActivityInfoResponse, err error)
	HCTest(ctx context.Context, req *HCTestRequest, opts ...http.CallOption) (rsp *HCTestResponse, err error)
	PreheatHelpCode(ctx context.Context, req *PreheatHelpCodeRequest, opts ...http.CallOption) (rsp *PreheatHelpCodeResponse, err error)
	RepairHelpCode(ctx context.Context, req *RepairHelpCodeRequest, opts ...http.CallOption) (rsp *RepairHelpCodeResponse, err error)
}

type HelpCodeHTTPClientImpl struct {
	cc *http.Client
}

func NewHelpCodeHTTPClient(client *http.Client) HelpCodeHTTPClient {
	return &HelpCodeHTTPClientImpl{client}
}

func (c *HelpCodeHTTPClientImpl) GetActivityInfo(ctx context.Context, in *GetActivityInfoRequest, opts ...http.CallOption) (*GetActivityInfoResponse, error) {
	var out GetActivityInfoResponse
	pattern := "/events/mlbb25031job/activity/getactivityinfo"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationHelpCodeGetActivityInfo))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *HelpCodeHTTPClientImpl) HCTest(ctx context.Context, in *HCTestRequest, opts ...http.CallOption) (*HCTestResponse, error) {
	var out HCTestResponse
	pattern := "/events/mlbb25031job/activity/helpCode/test"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationHelpCodeHCTest))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *HelpCodeHTTPClientImpl) PreheatHelpCode(ctx context.Context, in *PreheatHelpCodeRequest, opts ...http.CallOption) (*PreheatHelpCodeResponse, error) {
	var out PreheatHelpCodeResponse
	pattern := "/events/mlbb25031job/activity/helpCode/preHeat"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationHelpCodePreheatHelpCode))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *HelpCodeHTTPClientImpl) RepairHelpCode(ctx context.Context, in *RepairHelpCodeRequest, opts ...http.CallOption) (*RepairHelpCodeResponse, error) {
	var out RepairHelpCodeResponse
	pattern := "/events/mlbb25031job/activity/helpCode/repair"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationHelpCodeRepairHelpCode))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
