package posture_exception

import (
	"kubescape-config-service/dbhandler"
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
	postureException.GET("/:"+utils.GUID_FIELD, dbhandler.HandleGetDocWithGUIDInPath[*types.PostureExceptionPolicy])
	postureException.DELETE("/:"+utils.GUID_FIELD, dbhandler.HandleDeleteDoc)

	posturePost := postureException.Group("/")
	posturePost.Use(dbhandler.PostValidation[*types.PostureExceptionPolicy])
	posturePost.POST("/", dbhandler.HandlePostDocFromContext[*types.PostureExceptionPolicy])

	posturePut := postureException.Group("/")
	posturePut.Use(dbhandler.PutValidation[*types.PostureExceptionPolicy])
	posturePut.PUT("/", dbhandler.HandlePutDocFromContext[*types.PostureExceptionPolicy])
	posturePut.PUT("/:"+utils.GUID_FIELD, dbhandler.HandlePutDocFromContext[*types.PostureExceptionPolicy])

}
