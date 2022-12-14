package handlers

import (
	"config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

// ////////////////////////////////db handler middleware//////////////////////////////////
// DBContextMiddleware is a middleware that adds db parameters to the context
func DBContextMiddleware(collectionName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(consts.Collection, collectionName)
		c.Next()
	}
}
