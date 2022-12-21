package framework

import (
	"config-service/handlers"
	"config-service/types"
	"config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	handlers.AddRoutes(g, handlers.NewRouterOptionsBuilder[*types.Framework]().
		WithPath(consts.FrameworkPath).
		WithDBCollection(consts.FrameworkCollection).
		WithNameQuery(consts.FrameworkNameParam).
		WithDeleteByName(true).
		Get()...)
}
