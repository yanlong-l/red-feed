//go:build wireinject

package main

import (
	"red-feed/internal/events/article"
	"red-feed/internal/repository"
	"red-feed/internal/repository/cache"
	"red-feed/internal/repository/dao"
	"red-feed/internal/service"
	"red-feed/internal/web"
	ijwt "red-feed/internal/web/jwt"
	"red-feed/ioc"

	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,

		article.NewKafkaProducer,
		article.NewInteractiveReadEventBatchConsumer,

		// 初始化DAO层 和 Cache层
		dao.NewGORMUserDAO,
		dao.NewInteractiveDAO,
		dao.NewGORMArticleDao,
		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewRedisArticleCache,
		cache.NewRedisInteractiveCache,

		// 初始化Repo层
		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewArticleRepository,
		repository.NewInteractiveRepository,

		// 初始化Service层
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		service.NewInteractiveService,
		ioc.InitWechatService,
		ioc.InitSMSService,

		// web handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,

		ijwt.NewRedisJWTHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,

		ioc.InitLogger,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
