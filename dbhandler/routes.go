package dbhandler

import (
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

// AddPolicyRoutes adds common routes for policy
func AddPolicyRoutes[T types.DocContent](g *gin.Engine, path, dbCollection string, paramConf *scopeParamsConfig) {
	routerGroup := g.Group(path)
	routerGroup.Use(DBContextMiddleware(dbCollection))
	routerGroup.GET("/", HandleGetByQueryOrAll[T](consts.POLICY_NAME_PARAM, paramConf))
	routerGroup.GET("/:"+consts.GUID_FIELD, HandleGetDocWithGUIDInPath[T])
	routerGroup.POST("/", HandlePostDocWithValidation[T]()...)
	routerGroup.PUT("/", HandlePutDocWithValidation[T]()...)
	routerGroup.PUT("/:"+consts.GUID_FIELD, HandlePutDocWithValidation[T]()...)	
	routerGroup.DELETE("/", HandleDeleteDocByName[T](consts.POLICY_NAME_PARAM))
	routerGroup.DELETE("/:"+consts.GUID_FIELD,  HandleDeleteDoc[T])
}
