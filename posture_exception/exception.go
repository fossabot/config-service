package posture_exception

import (
	"fmt"
	"kubescape-config-service/mongo"
	"kubescape-config-service/types"
	"kubescape-config-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getPostureExceptionPolicies(c *gin.Context) {
	if _, list := c.GetQuery("list"); list {
		//get all policies names
		namesProjection := mongo.NewProjectionBuilder().Include("name").Get()
		if policiesNames, err := mongo.GetAllForCustomerWithProjection(c, []string{}, namesProjection); err != nil {
			utils.LogNTraceError("failed to read polices", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, policiesNames)
			return
		}
	} else if policyName := c.Query("policyName"); policyName != "" {
		//get policy by name
		if policy, err := mongo.GetDocByName(c, policyName, &types.PostureExceptionPolicy{}); err != nil {
			utils.LogNTraceError("failed to read policy", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, policy)
			return
		}
	}

	//get all policies
	if policies, err := mongo.GetAllForCustomer(c, []types.PostureExceptionPolicy{}); err != nil {
		utils.LogNTraceError("failed to read policys", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, policies)
	}
}

/*
func getPolicy(c *gin.Context) {
	guid := c.Param(utils.GUID_FIELD)
	if guid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "policy guid is required"})
		return
	}

	if policy, err := mongo.GetDocByGUID(c, guid, &types.policy{}); err != nil {
		utils.LogNTraceError("failed to read policy", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, policy)
	}
}*/

func postPostureExceptionPolicy(c *gin.Context) {
	reqPolicy := types.PostureExceptionPolicy{}
	if err := c.BindJSON(&reqPolicy); err != nil {
		return
	}
	if reqPolicy.Name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "policy name is required"})
		return
	}
	if exist, err := mongo.DocWithNameExist(c, reqPolicy.Name); err != nil {
		utils.LogNTraceError("failed to check if policy name exist", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("policy with name %s already exists", reqPolicy.Name)})
		return
	}

	policyDoc := mongo.NewDocument(&reqPolicy, c.GetString(utils.CUSTOMER_GUID))
	if result, err := mongo.GetWriteCollection(utils.POSTURE_EXCEPTION_POLICIES).InsertOne(c.Request.Context(), policyDoc); err != nil {
		utils.LogNTraceError("failed to create policy", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"GUID": result.InsertedID})
	}
}

func putPostureExceptionPolicy(c *gin.Context) {
	reqPolicy := types.PostureExceptionPolicy{}
	if err := c.BindJSON(&reqPolicy); err != nil {
		return
	}
	if guid := c.Param(utils.GUID_FIELD); guid != "" {
		reqPolicy.GUID = guid
	}
	if reqPolicy.GUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "policy guid is required"})
		return
	}

	update, err := mongo.GetUpdateDocCommand(&reqPolicy)
	if err != nil {
		utils.LogNTraceError("failed to create update command", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.LogNTrace(fmt.Sprintf("post policy %s - updating policy", reqPolicy.GUID), c)
	if updatedPolicy, err := mongo.UpdateDocument(c, reqPolicy.GUID, update, &types.PostureExceptionPolicy{}); err != nil {
		utils.LogNTraceError("failed to update policy", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, updatedPolicy)
	}

}
