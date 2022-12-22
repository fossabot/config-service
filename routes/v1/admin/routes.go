package admin

import (
	"config-service/db"
	"config-service/handlers"
	"config-service/utils"
	"config-service/utils/consts"
	"config-service/utils/log"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
)

func AddRoutes(g *gin.Engine) {
	admin := g.Group(consts.AdminPath)

	//add middleware to check if user is admin
	adminUsers := utils.GetConfig().AdminUsers
	adminAuthMiddleware := func(c *gin.Context) {
		if slices.Contains(adminUsers, c.GetString(consts.CustomerGUID)) {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - not an admin user"})
		}
	}
	admin.Use(adminAuthMiddleware)

	//add delete customers data route
	admin.DELETE("customers", deleteAllCustomerData)
}

func deleteAllCustomerData(c *gin.Context) {
	customersGUIDs := c.QueryArray(consts.CustomersParam)
	if len(customersGUIDs) == 0 {
		handlers.ResponseBadRequest(c, consts.CustomersParam+" query param is required")
		return
	}
	deleted, err := db.AdminDeleteCustomersDocs(c, customersGUIDs...)
	if err != nil {
		log.LogNTraceError(fmt.Sprintf("deleteAllCustomerData ended with errors. %d documents deleted", deleted), err, c)
		handlers.ResponseInternalServerError(c, fmt.Sprintf("deleted: %d, errors: %v", deleted, err), err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": deleted})

}
