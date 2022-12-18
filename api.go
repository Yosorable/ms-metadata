package main

import (
	"context"

	"github.com/Yosorable/ms-metadata/core/handler"
	pb "github.com/Yosorable/ms-shared/protoc_gen/metadata"
)

type metadataServer struct {
	pb.UnimplementedMetadataServer
}

func (metadataServer) CreateObj(ctx context.Context, req *pb.CreateObjRequest) (*pb.CreateObjReply, error) {
	return handler.CreateObj(ctx, req)
}
