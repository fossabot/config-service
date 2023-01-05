package registry_cron_job

import (
	"config-service/handlers"
	"config-service/types"
	"config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	handlers.AddRoutes(g, handlers.NewRouterOptionsBuilder[*types.RegistryCronJob]().
		WithPath(consts.RegistryCronJobPath).
		WithDBCollection(consts.RegistryCronJobCollection).
		WithValidatePostUniqueName(true).
		WithValidatePutGUID(true).
		WithDeleteByName(true).
		WithNameQuery(consts.NameField).
		WithQueryConfig(handlers.FlatQueryConfig()).
		Get()...)
}
