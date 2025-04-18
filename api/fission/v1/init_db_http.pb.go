// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.8.3
// - protoc             v5.29.3
// source: fission/v1/init_db.proto

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

const OperationInitDBInitDB = "/fission.v1.InitDB/InitDB"
const OperationInitDBQuerySql = "/fission.v1.InitDB/QuerySql"

type InitDBHTTPServer interface {
	InitDB(context.Context, *InitDBRequest) (*InitDBRequestResponse, error)
	QuerySql(context.Context, *QuerySqlRequest) (*QuerySqlResponse, error)
}

func RegisterInitDBHTTPServer(s *http.Server, srv InitDBHTTPServer) {
	r := s.Route("/")
	r.POST("/events/mlbb25031gateway/activity/initDB", _InitDB_InitDB0_HTTP_Handler(srv))
	r.POST("/events/mlbb25031gateway/activity/sql-query", _InitDB_QuerySql0_HTTP_Handler(srv))
}

func _InitDB_InitDB0_HTTP_Handler(srv InitDBHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in InitDBRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationInitDBInitDB)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.InitDB(ctx, req.(*InitDBRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*InitDBRequestResponse)
		return ctx.Result(200, reply)
	}
}

func _InitDB_QuerySql0_HTTP_Handler(srv InitDBHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in QuerySqlRequest
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationInitDBQuerySql)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.QuerySql(ctx, req.(*QuerySqlRequest))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*QuerySqlResponse)
		return ctx.Result(200, reply)
	}
}

type InitDBHTTPClient interface {
	InitDB(ctx context.Context, req *InitDBRequest, opts ...http.CallOption) (rsp *InitDBRequestResponse, err error)
	QuerySql(ctx context.Context, req *QuerySqlRequest, opts ...http.CallOption) (rsp *QuerySqlResponse, err error)
}

type InitDBHTTPClientImpl struct {
	cc *http.Client
}

func NewInitDBHTTPClient(client *http.Client) InitDBHTTPClient {
	return &InitDBHTTPClientImpl{client}
}

func (c *InitDBHTTPClientImpl) InitDB(ctx context.Context, in *InitDBRequest, opts ...http.CallOption) (*InitDBRequestResponse, error) {
	var out InitDBRequestResponse
	pattern := "/events/mlbb25031gateway/activity/initDB"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationInitDBInitDB))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *InitDBHTTPClientImpl) QuerySql(ctx context.Context, in *QuerySqlRequest, opts ...http.CallOption) (*QuerySqlResponse, error) {
	var out QuerySqlResponse
	pattern := "/events/mlbb25031gateway/activity/sql-query"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationInitDBQuerySql))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
