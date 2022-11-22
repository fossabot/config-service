package login

import (
	"kubescape-config-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	login := g.Group("/login")

	//login routes
	login.POST("/", func(c *gin.Context) {
		loginDetails := struct {
			CustomerGUID string `json:"customerGUID" binding:"required"`
		}{
			CustomerGUID: "",
		}

		if err := c.BindJSON(&loginDetails); err != nil {
			return
		}
		c.SetCookie(utils.CUSTOMER_GUID, loginDetails.CustomerGUID, 2*60*60*24, "/", "", false, true)
		c.JSON(http.StatusOK, nil)
	})
}
