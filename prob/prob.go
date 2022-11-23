package prob

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	prob := g.Group("/")

	prob.GET("liveliness", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	prob.GET("readiness", func(c *gin.Context) {
		//TODO check number of open mongo sessions
		c.JSON(http.StatusOK, nil)
	})

}
