//go:build wireinject

package main

import (
	"github.com/google/wire"
	"red-feed/interactive/events"
	"red-feed/interactive/grpc"
	"red-feed/interactive/ioc"
	"red-feed/interactive/repository"
	"red-feed/interactive/repository/cache"
	"red-feed/interactive/repository/dao"
	"red-feed/interactive/service"
)

var thirdPartySet = wire.NewSet(ioc.InitDB,
	ioc.InitLogger,
	ioc.InitKafka,
	// 暂时不理会 consumer 怎么启动
	ioc.InitRedis)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewInteractiveRepository,
	dao.NewInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitAPP() *App {
	wire.Build(interactiveSvcProvider,
		thirdPartySet,
		events.NewInteractiveReadEventConsumer,
		grpc.NewInteractiveServiceServer,
		ioc.NewConsumers,
		ioc.InitGRPCXServer,
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
