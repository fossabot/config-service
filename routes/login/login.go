package login

import (
	"config-service/handlers"
	"config-service/utils/consts"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	login := g.Group("/login")

	//login routes
	login.POST("", func(c *gin.Context) {
		loginDetails := struct {
			CustomerGUID string `json:"customerGUID" binding:"required"`
		}{
			CustomerGUID: "",
		}

		if err := c.ShouldBindJSON(&loginDetails); err != nil {
			handlers.ResponseFailedToBindJson(c, err)
			return
		}
		c.SetCookie(consts.CustomerGUID, loginDetails.CustomerGUID, 2*60*60*24, "/", "", false, true)
		c.JSON(http.StatusOK, nil)
	})
}
