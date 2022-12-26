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
		//check if admin access granted by auth middleware or if user is in the configuration admin users list
		if c.GetBool(consts.AdminAccess) {
			c.Next()
		} else if slices.Contains(adminUsers, c.GetString(consts.CustomerGUID)) {
			c.Next()
		} else {
			//not admin
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized - not an admin user"})
		}
	}

	admin.Use(adminAuthMiddleware)

	//add delete customers data route
	admin.DELETE("/customers", deleteAllCustomerData)
}

func deleteAllCustomerData(c *gin.Context) {
	customersGUIDs := c.QueryArray(consts.CustomersParam)
	if len(customersGUIDs) == 0 {
		handlers.ResponseMissingQueryParam(c, consts.CustomersParam)
		return
	}
	deleted, err := db.AdminDeleteCustomersDocs(c, customersGUIDs...)
	if err != nil {
		log.LogNTraceError(fmt.Sprintf("deleteAllCustomerData completed with errors. %d documents deleted", deleted), err, c)
		handlers.ResponseInternalServerError(c, fmt.Sprintf("deleted: %d, errors: %v", deleted, err), err)
		return
	}
	log.LogNTrace(fmt.Sprintf("deleteAllCustomerData completed successfully. %d documents of %d users deleted by admin %s ", deleted, len(customersGUIDs), c.GetString(consts.CustomerGUID)), c)
	c.JSON(http.StatusOK, gin.H{"deleted": deleted})

}
