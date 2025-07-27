package ioc

import (
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	igrpc "red-feed/interactive/grpc"
	"red-feed/pkg/grpcx"
)

func InitGRPCXServer(intrGRPCServer *igrpc.InteractiveServiceServer) *grpcx.Server {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.server", &cfg)
	if err != nil {
		panic(err)
	}

	server := grpc.NewServer()
	intrGRPCServer.Register(server)
	return &grpcx.Server{
		Server: server,
		Addr:   cfg.Addr,
	}
}
