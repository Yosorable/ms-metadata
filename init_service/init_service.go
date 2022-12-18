package init_service

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/Yosorable/ms-metadata/global"
	"github.com/Yosorable/ms-shared/utils"
	"github.com/Yosorable/ms-shared/utils/database"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
)

func InitService() *grpc.Server {
	if err := utils.LoadJsonConfigFileWithDefaultPath(&global.CONFIG); err != nil {
		panic(err)
	}

	conf := global.CONFIG

	if db, err := database.InitMysql(
		conf.MySQL.User,
		conf.MySQL.Password,
		conf.MySQL.Addr,
		conf.MySQL.DBName,
	); err != nil {
		panic(err)
	} else {
		global.DATABASE = db
	}

	return grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(
				grpc_recovery.WithRecoveryHandler(func(p any) (err error) {
					log.Printf("[panic] %v\n %s\n", p, debug.Stack())
					return fmt.Errorf("%v", p)
				}),
			),
		)),
	)
}
