//go:build wireinject

package startup

import (
	"github.com/google/wire"
	"red-feed/interactive/repository"
	"red-feed/interactive/repository/cache"
	"red-feed/interactive/repository/dao"
	"red-feed/interactive/service"
)

var thirdProvider = wire.NewSet(InitRedis,
	InitTestDB, InitTestLogger)

var interactiveSvcProvider = wire.NewSet(
	service.NewInteractiveService,
	repository.NewInteractiveRepository,
	dao.NewInteractiveDAO,
	cache.NewRedisInteractiveCache,
)

func InitInteractiveService() service.InteractiveService {
	wire.Build(thirdProvider, interactiveSvcProvider)
	return service.NewInteractiveService(nil, nil)
}
