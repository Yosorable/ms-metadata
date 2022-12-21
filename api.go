package main

import (
	"context"

	"github.com/Yosorable/ms-metadata/core/handler"
	common_pb "github.com/Yosorable/ms-shared/protoc_gen/common"
	pb "github.com/Yosorable/ms-shared/protoc_gen/metadata"
	mgrpc "github.com/Yosorable/ms-shared/utils/grpc"
)

type metadataServer struct {
	pb.UnimplementedMetadataServer
}

func (mts metadataServer) DynamicCall(ctx context.Context, req *common_pb.DynamicCallRequest) (*common_pb.DynamicCallReply, error) {
	return mgrpc.DynamicCall(ctx, req, &mts)
}

func (metadataServer) CreateObj(ctx context.Context, req *pb.CreateObjRequest) (*pb.CreateObjReply, error) {
	return handler.CreateObj(ctx, req)
}

func (metadataServer) GetObjByID(ctx context.Context, req *pb.GetObjRequest) (*pb.GetObjByIDReply, error) {
	return handler.GetObjByID(ctx, req)
}

func (metadataServer) GetObjByName(ctx context.Context, req *pb.GetObjRequest) (*pb.GetObjByNameReply, error) {
	return handler.GetObjByName(ctx, req)
}
