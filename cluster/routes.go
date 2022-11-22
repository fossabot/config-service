package cluster

import (
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	cluster := g.Group("/cluster")

	cluster.Use(dbhandler.DBContextMiddleware(consts.CLUSTERS))

	cluster.GET("/", dbhandler.HandleGetAll[*types.Cluster])
	cluster.GET("/:"+consts.GUID_FIELD, dbhandler.HandleGetDocWithGUIDInPath[*types.Cluster])
	cluster.POST("/", dbhandler.PostValidation[*types.Cluster], postCluster)
	cluster.PUT("/", dbhandler.PutValidation[*types.Cluster], putCluster)
	cluster.PUT("/:"+consts.GUID_FIELD, putCluster)
	cluster.DELETE("/:"+consts.GUID_FIELD, dbhandler.HandleDeleteDoc)
}
