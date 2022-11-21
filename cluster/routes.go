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
	cluster.DELETE("/:"+utils.GUID_FIELD, dbhandler.HandleDeleteDoc)

	clusterPost := cluster.Group("/")
	clusterPost.Use(dbhandler.PostValidation[*types.Cluster])
	clusterPost.POST("/", postCluster)

	clusterPut := cluster.Group("/")
	clusterPut.Use(dbhandler.PutValidation[*types.Cluster])
	clusterPut.PUT("/", putCluster)
	clusterPut.PUT("/:"+utils.GUID_FIELD, putCluster)

}
