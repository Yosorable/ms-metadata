package main

import (
	"github.com/Yosorable/ms-metadata/global"
	"github.com/Yosorable/ms-metadata/init_service"
	pb "github.com/Yosorable/ms-shared/protoc_gen/metadata"
	mgrpc "github.com/Yosorable/ms-shared/utils/grpc"
)

func main() {
	grpcServer := init_service.InitService()

	pb.RegisterMetadataServer(grpcServer, &metadataServer{})

	mgrpc.RunRpcServerInLocalHost(
		global.CONFIG.ServiceName,
		grpcServer,
	)
}
