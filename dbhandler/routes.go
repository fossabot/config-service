package dbhandler

import (
	"config-service/types"
	"config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

// AddPolicyRoutes adds common routes for policy
func AddPolicyRoutes[T types.DocContent](g *gin.Engine, path, dbCollection string, paramConf *scopeParamsConfig) {
	routerGroup := g.Group(path)
	routerGroup.Use(DBContextMiddleware(dbCollection))
	routerGroup.GET("", HandleGetByQueryOrAll[T](consts.PolicyNameParam, paramConf, true))
	routerGroup.GET("/:"+consts.GUIDField, HandleGetDocWithGUIDInPath[T])
	routerGroup.POST("", HandlePostDocWithValidation[T]()...)
	routerGroup.PUT("", HandlePutDocWithValidation[T]()...)
	routerGroup.PUT("/:"+consts.GUIDField, HandlePutDocWithValidation[T]()...)
	routerGroup.DELETE("", HandleDeleteDocByName[T](consts.PolicyNameParam))
	routerGroup.DELETE("/:"+consts.GUIDField, HandleDeleteDoc[T])
}
