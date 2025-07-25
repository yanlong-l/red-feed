//go:build wireinject

package main

import (
	"red-feed/interactive/events"
	repository2 "red-feed/interactive/repository"
	cache2 "red-feed/interactive/repository/cache"
	dao2 "red-feed/interactive/repository/dao"
	service2 "red-feed/interactive/service"
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

var rankingServiceSet = wire.NewSet(
	repository.NewCachedRankingRepository,
	cache.NewRankingRedisCache,
	service.NewBatchRankingService,
)

func InitApp() *App {
	wire.Build(
		// 最基础的第三方依赖
		ioc.InitRedis, ioc.InitDB,
		ioc.InitKafka,
		ioc.NewConsumers,
		ioc.NewSyncProducer,

		article.NewKafkaProducer,
		events.NewInteractiveReadEventBatchConsumer,

		// 初始化DAO层 和 Cache层
		dao.NewGORMUserDAO,
		dao2.NewInteractiveDAO,
		dao.NewGORMArticleDao,
		cache.NewUserCache,
		cache.NewCodeCache,
		cache.NewRedisArticleCache,
		cache2.NewRedisInteractiveCache,

		// 初始化Repo层
		repository.NewUserRepository,
		repository.NewCodeRepository,
		repository.NewArticleRepository,
		repository2.NewInteractiveRepository,

		// 初始化Service层
		service.NewUserService,
		service.NewCodeService,
		service.NewArticleService,
		service2.NewInteractiveService,
		ioc.InitWechatService,
		ioc.InitSMSService,

		// web handler
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,

		ijwt.NewRedisJWTHandler,
		ioc.InitMiddlewares,
		ioc.InitWebServer,

		rankingServiceSet,
		ioc.InitJobs,
		ioc.InitRankingJob,
		ioc.InitRLockClient,

		ioc.InitLogger,

		wire.Struct(new(App), "*"),
	)
	return new(App)
}
