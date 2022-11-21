package posture_exception

import (
	"fmt"
	"kubescape-config-service/mongo"
	"kubescape-config-service/utils"
	"net/http"

	"github.com/armosec/armoapi-go/armotypes"
	"github.com/gin-gonic/gin"
)

func getPostureExceptionPolicies(c *gin.Context) {
	if _, list := c.GetQuery("list"); list {
		//get all policies names
		namesProjection := mongo.NewProjectionBuilder().Include("name").Get()
		if policiesNames, err := mongo.GetAllForCustomerWithProjection(c, []string{}, namesProjection); err != nil {
			utils.LogNTraceError("failed to read clusters", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, policiesNames)
			return
		}
	} else if policyName := c.Query("policyName"); policyName != "" {
		//get policy by name
		if policy, err := mongo.GetDocByName(c, policyName, &armotypes.PostureExceptionPolicy{}); err != nil {
			utils.LogNTraceError("failed to read policy", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(http.StatusOK, policy)
			return
		}
	}

	//get all policies
	if policies, err := mongo.GetAllForCustomer(c, []armotypes.PostureExceptionPolicy{}); err != nil {
		utils.LogNTraceError("failed to read clusters", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, policies)
	}
}

func postPostureExceptionPolicy(c *gin.Context) {
	reqPolicy := armotypes.PostureExceptionPolicy{}
	if err := c.ShouldBindJSON(&reqPolicy); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	policyDoc := mongo.NewPostureExceptionDocument(reqPolicy, c.GetString(utils.CUSTOMER_GUID))
	if result, err := mongo.GetWriteCollection(utils.POSTURE_EXCEPTION_POLICIES).InsertOne(c.Request.Context(), policyDoc); err != nil {
		utils.LogNTraceError("failed to create policy", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"GUID": result.InsertedID})
	}
}
