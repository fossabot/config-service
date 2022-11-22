package cluster

import (
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	cluster := g.Group("/cluster")

	cluster.Use(func(c *gin.Context) {
		//set clusters collection name in context - used by mongo utils functions
		c.Set(utils.COLLECTION, utils.CLUSTERS)
		c.Next()
	})
	cluster.GET("/", dbhandler.HandleGetAll[*types.Cluster])
	cluster.GET("/:"+utils.GUID_FIELD, dbhandler.HandleGetDocWithGUIDInPath[*types.Cluster])
	cluster.POST("/", dbhandler.PostValidation[*types.Cluster], postCluster)
	cluster.PUT("/", dbhandler.PutValidation[*types.Cluster], putCluster)
	cluster.PUT("/:"+utils.GUID_FIELD, putCluster)
	cluster.DELETE("/:"+utils.GUID_FIELD, dbhandler.HandleDeleteDoc)
}
