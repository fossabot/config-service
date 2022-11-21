package cluster

import (
	"kubescape-config-service/mongo"
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

	cluster.GET("/", mongo.HandleGetAll[*types.Cluster])
	cluster.GET("/:"+utils.GUID_FIELD, mongo.HandleGetDocWithGUIDInPath[*types.Cluster])
	cluster.DELETE("/:"+utils.GUID_FIELD, mongo.HandleDeleteDoc)

	clusterPost := cluster.Group("/")
	clusterPost.Use(mongo.PostValidation[*types.Cluster])
	clusterPost.POST("/", postCluster)

	clusterPut := cluster.Group("/")
	clusterPut.Use(mongo.PutValidation[*types.Cluster])
	clusterPut.PUT("/", putCluster)
	clusterPut.PUT("/:"+utils.GUID_FIELD, putCluster)

}
