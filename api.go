package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/Yosorable/ms-metadata/core/handler"
	common_pb "github.com/Yosorable/ms-shared/protoc_gen/common"
	pb "github.com/Yosorable/ms-shared/protoc_gen/metadata"
	"github.com/Yosorable/ms-shared/utils"
)

type metadataServer struct {
	pb.UnimplementedMetadataServer
}

func (mts metadataServer) DynamicCall(ctx context.Context, req *common_pb.DynamicCallRequest) (*common_pb.DynamicCallReply, error) {
	if req.GetMethod() == "" || req.GetMethod() == "DynamicCall" {
		return nil, utils.NewStatusError(7000, "please set correct method name")
	}

	serverRef := reflect.ValueOf(mts)
	methodRef := serverRef.MethodByName(req.GetMethod())

	if methodRef.Kind() == 0 {
		return nil, utils.NewStatusError(7000, fmt.Sprintf("method [%s] not implement", req.GetMethod()))
	}

	ctxRef := reflect.ValueOf(ctx)
	reqRefTp := methodRef.Type().In(1)
	reqRef := reflect.New(reqRefTp)
	if jsonErr := json.Unmarshal([]byte(req.GetData()), reqRef.Interface()); jsonErr != nil {
		return nil, utils.NewStatusError(7000, jsonErr)
	}
	reqRef = reqRef.Elem()

	resRefSlice := methodRef.Call([]reflect.Value{
		ctxRef,
		reqRef,
	})

	resRef, errRef := resRefSlice[0], resRefSlice[1]
	if !errRef.IsNil() {
		return nil, errRef.Interface().(error)
	}

	resData, err := json.Marshal(resRef.Interface())
	if err != nil {
		return nil, utils.NewStatusError(7000, err)
	}

	return &common_pb.DynamicCallReply{
		Data: string(resData),
	}, nil
}

func (metadataServer) CreateObj(ctx context.Context, req *pb.CreateObjRequest) (*pb.CreateObjReply, error) {
	return handler.CreateObj(ctx, req)
}
