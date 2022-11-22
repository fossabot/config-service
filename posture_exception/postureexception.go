package posture_exception

import (
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"
	"kubescape-config-service/utils/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getPostureExceptionPolicies(c *gin.Context) {
	if _, list := c.GetQuery("list"); list {
		namesProjection := dbhandler.NewProjectionBuilder().Include(consts.NAME_FIELD).ExcludeID().Get()
		if policiesNames, err := dbhandler.GetAllForCustomerWithProjection[types.PostureExceptionPolicy](c, namesProjection); err != nil {
			log.LogNTraceError("failed to read polices", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			var names []string
			for _, policy := range policiesNames {
				names = append(names, policy.Name)
			}
			c.JSON(http.StatusOK, names)
			return
		}
	} else if policyName := c.Query("policyName"); policyName != "" {
		//get policy by name
		if policy, err := dbhandler.GetDocByName[types.PostureExceptionPolicy](c, policyName); err != nil {
			log.LogNTraceError("failed to read policy", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, policy)
			return
		}
	}
	dbhandler.HandleGetAll[*types.PostureExceptionPolicy](c)
}
