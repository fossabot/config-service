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
			CustomerGUID string                 `json:"customerGUID" binding:"required"`
			Attributes   map[string]interface{} `json:"attributes,omitempty"`
		}{
			CustomerGUID: "",
		}

		if err := c.ShouldBindJSON(&loginDetails); err != nil {
			handlers.ResponseFailedToBindJson(c, err)
			return
		}
		if loginDetails.CustomerGUID == "" {
			handlers.ResponseBadRequest(c, "customerGUID is required")
			return
		}
		cookieValue := loginDetails.CustomerGUID
		//check if admin access required
		if loginDetails.Attributes != nil {
			if admin, ok := loginDetails.Attributes["admin"]; ok && admin == true {
				cookieValue += ";" + consts.AdminAccess
			}
		}
		c.SetCookie(consts.CustomerGUID, cookieValue, 2*60*60*24, "/", "", false, true)
		c.JSON(http.StatusOK, nil)
	})
}
