package prob

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {

	g.GET("/liveliness", func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
	})

	g.GET("/readiness", func(c *gin.Context) {
		//TODO check number of openn mongo sessions
		c.JSON(http.StatusOK, nil)
	})

}
