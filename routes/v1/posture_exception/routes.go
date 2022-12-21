package posture_exception

import (
	"config-service/handlers"
	"config-service/types"
	"config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	handlers.AddPolicyRoutes[*types.PostureExceptionPolicy](g,
		consts.PostureExceptionPolicyPath,
		consts.PostureExceptionPolicyCollection, handlers.GetPostureExceptionQueryConfig())
}
