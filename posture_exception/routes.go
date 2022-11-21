package posture_exception

import (
	"kubescape-config-service/mongo"
	"kubescape-config-service/types"
	"kubescape-config-service/utils"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	postureException := g.Group("/v1_posture_exception_policy")

	postureException.Use(func(c *gin.Context) {
		//set posture_exception collection name in context - used by mongo utils functions
		c.Set(utils.COLLECTION, utils.POSTURE_EXCEPTION_POLICIES)
		c.Next()
	})

	postureException.GET("/", getPostureExceptionPolicies)
	postureException.GET("/:"+utils.GUID_FIELD, mongo.HandleGetDocWithGUIDInPath[*types.PostureExceptionPolicy])
	postureException.DELETE("/:"+utils.GUID_FIELD, mongo.HandleDeleteDoc)

	posturePost := postureException.Group("/")
	posturePost.Use(mongo.PostValidation[*types.PostureExceptionPolicy])
	posturePost.POST("/", mongo.HandlePostDocFromContext[*types.PostureExceptionPolicy])

	posturePut := postureException.Group("/")
	posturePut.Use(mongo.PutValidation[*types.PostureExceptionPolicy])
	posturePut.PUT("/", mongo.HandlePutDocFromContext[*types.PostureExceptionPolicy])
	posturePut.PUT("/:"+utils.GUID_FIELD, mongo.HandlePutDocFromContext[*types.PostureExceptionPolicy])

}
