package cluster

import (
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	cluster := g.Group(consts.ClusterPath)

	cluster.Use(dbhandler.DBContextMiddleware(consts.ClustersCollection))

	cluster.GET("", dbhandler.HandleGetAll[*types.Cluster])
	cluster.GET("/:"+consts.GUIDField, dbhandler.HandleGetDocWithGUIDInPath[*types.Cluster])
	cluster.POST("", dbhandler.HandlePostValidation[*types.Cluster], postCluster)
	cluster.PUT("", dbhandler.HandlePutValidation[*types.Cluster], putCluster)
	cluster.PUT("/:"+consts.GUIDField, dbhandler.HandlePutValidation[*types.Cluster], putCluster)
	cluster.DELETE("/:"+consts.GUIDField, dbhandler.HandleDeleteDoc[*types.Cluster])
}
