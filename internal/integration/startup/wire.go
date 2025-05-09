//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"red-feed/internal/repository"
	"red-feed/internal/repository/dao"
	"red-feed/internal/service"
	"red-feed/internal/web"
)

var thirdProvider = wire.NewSet(InitRedis, InitTestDB, InitTestLogger)

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdProvider,
		service.NewArticleService,
		web.NewArticleHandler,
		repository.NewArticleRepository,
		dao.NewGORMArticleDao,
	)
	return &web.ArticleHandler{}
}
