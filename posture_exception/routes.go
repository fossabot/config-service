package posture_exception

import (
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	postureException := g.Group("/v1_posture_exception_policy")

	postureException.Use(dbhandler.DBContextMiddleware(consts.POSTURE_EXCEPTION_POLICIES))

	postureException.GET("/", getPostureExceptionPolicies)
	postureException.GET("/:"+consts.GUID_FIELD, dbhandler.HandleGetDocWithGUIDInPath[*types.PostureExceptionPolicy])
	postureException.POST("/", dbhandler.HandlePostDocWithValidation[*types.PostureExceptionPolicy]()...)
	postureException.PUT("/", dbhandler.HandlePutDocWithValidation[*types.PostureExceptionPolicy]()...)
	postureException.PUT("/:"+consts.GUID_FIELD, dbhandler.HandlePutDocWithValidation[*types.PostureExceptionPolicy]()...)
	postureException.DELETE("/:"+consts.GUID_FIELD, dbhandler.HandleDeleteDoc)

}
