package posture_exception

import (
	"config-service/dbhandler"
	"config-service/types"
	"config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	dbhandler.AddPolicyRoutes[*types.PostureExceptionPolicy](g,
		consts.PostureExceptionPolicyPath,
		consts.PostureExceptionPolicyCollection, dbhandler.GetPostureExceptionQueryConfig())
}
