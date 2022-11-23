package posture_exception

import (
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	dbhandler.AddPolicyRoutes[*types.PostureExceptionPolicy](g,
		consts.POSTURE_EXCEPTION_POLICY_PATH,
		consts.POSTURE_EXCEPTION_POLICIES_COLLECTION, dbhandler.GetPostureExceptionQueryConfig())
}
