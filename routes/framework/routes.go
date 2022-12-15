package framework

import (
	"config-service/handlers"
	"config-service/types"
	"config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	handlers.AddRoutes(g,
		handlers.WithPath[*types.Framework](consts.FrameworkPath),
		handlers.WithDBCollection[*types.Framework](consts.FrameworkCollection),
		handlers.WithNameQuery[*types.Framework](consts.FrameworkNameParam),
		handlers.WithDeleteByName[*types.Framework](true),
	)
}
