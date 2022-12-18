package main

import (
	"context"

	pb "github.com/Yosorable/ms-shared/protoc_gen/metadata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type metadataServer struct {
	pb.UnimplementedMetadataServer
}

func (metadataServer) SayHello(context.Context, *pb.SayHelloRequest) (*pb.SayHelloReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SayHello not implemented")
}
