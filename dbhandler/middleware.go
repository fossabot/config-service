package dbhandler

import (
	"kubescape-config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

func DBContextMiddleware(collectionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//set collection name in context - used by db handlers
		c.Set(consts.COLLECTION, collectionName)
		c.Next()
	}
}
