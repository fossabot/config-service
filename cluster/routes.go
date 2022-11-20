package cluster

import (
	"kubescape-config-service/mongo"
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

	cluster.GET("/", getClusters)

	cluster.GET("/:"+utils.GUID_FIELD, getCluster)

	cluster.POST("/", postCluster)

	cluster.PUT("/", putCluster)

	cluster.PUT("/:"+utils.GUID_FIELD, putCluster)

	cluster.DELETE("/:"+utils.GUID_FIELD, mongo.HandleDeleteDoc)
}
