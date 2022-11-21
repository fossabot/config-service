package posture_exception

import (
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getPostureExceptionPolicies(c *gin.Context) {
	if _, list := c.GetQuery("list"); list {
		//get all policies names
		namesProjection := dbhandler.NewProjectionBuilder().Include("name").Get()
		if policiesNames, err := dbhandler.GetAllForCustomerWithProjection(c, []string{}, namesProjection); err != nil {
			utils.LogNTraceError("failed to read polices", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, policiesNames)
			return
		}
	} else if policyName := c.Query("policyName"); policyName != "" {
		//get policy by name
		if policy, err := dbhandler.GetDocByName(c, policyName, &types.PostureExceptionPolicy{}); err != nil {
			utils.LogNTraceError("failed to read policy", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, policy)
			return
		}
	}
	dbhandler.HandleGetAll[*types.PostureExceptionPolicy](c)
}
