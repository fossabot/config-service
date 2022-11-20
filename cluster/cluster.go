package cluster

import (
	"fmt"
	"kubescape-config-service/mongo"
	"kubescape-config-service/types"
	"kubescape-config-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func getClusters(c *gin.Context) {
	if clusters, err := mongo.GetAllForCustomer(c, []types.Cluster{}); err != nil {
		utils.LogNTraceError("failed to read clusters", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, clusters)
	}
}

func getCluster(c *gin.Context) {
	guid := c.Param(utils.GUID_FIELD)
	if guid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cluster guid is required"})
		return
	}

	if cluster, err := mongo.GetDocByGUID(c, guid, &types.Cluster{}); err != nil {
		utils.LogNTraceError("failed to read cluster", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, cluster)
	}
}

func postCluster(c *gin.Context) {
	reqCluster := types.Cluster{}
	if err := c.ShouldBindJSON(&reqCluster); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if reqCluster.Name == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cluster name is required"})
		return
	}
	if exist, err := mongo.DocExist(c,
		mongo.NewFilterBuilder().
			WithValue("name", reqCluster.Name).
			Get()); err != nil {
		utils.LogNTraceError("failed to check if cluster name exist", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("cluster with name %s already exists", reqCluster.Name)})
		return
	}

	clusterDoc := mongo.NewClusterDocument(reqCluster)
	clusterDoc.Customers = append(clusterDoc.Customers, c.GetString(utils.CUSTOMER_GUID))
	clusterDoc.Attributes[utils.SHORT_NAME_ATTRIBUTE] = getUniqueShortName(clusterDoc.Name, c)

	if result, err := mongo.GetWriteCollection(utils.CLUSTERS).InsertOne(c.Request.Context(), clusterDoc); err != nil {
		utils.LogNTraceError("failed to create cluster", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"GUID": result.InsertedID})
	}
}

func putCluster(c *gin.Context) {
	reqCluster := types.Cluster{}
	if err := c.ShouldBindJSON(&reqCluster); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if guid := c.Param(utils.GUID_FIELD); guid != "" {
		reqCluster.GUID = guid
	}
	if reqCluster.GUID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cluster guid is required"})
		return
	}
	if reqCluster.Attributes == nil {
		reqCluster.Attributes = map[string]interface{}{}
	}
	// if request does attributes does not include alias add if from the old cluster
	if _, ok := reqCluster.Attributes[utils.SHORT_NAME_ATTRIBUTE]; !ok {
		if oldCluster, err := mongo.GetDocByGUID(c, reqCluster.GUID, &types.Cluster{}); err != nil {
			utils.LogNTraceError("failed to read cluster", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			reqCluster.Attributes[utils.SHORT_NAME_ATTRIBUTE] = oldCluster.Attributes[utils.SHORT_NAME_ATTRIBUTE]
		}
	}
	//only attributes can be updated- so check if there are any attributes
	if len(reqCluster.Attributes) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cluster attributes are required"})
		return
	}

	update := mongo.GetSetUpdate(reqCluster.Attributes, utils.ATTRIBUTES_FIELD)
	utils.LogNTrace(fmt.Sprintf("post cluster %s - updating cluster", reqCluster.GUID), c)
	if updatedCluster, err := mongo.UpdateDocument(c, reqCluster.GUID, update, &types.Cluster{}); err != nil {
		utils.LogNTraceError("failed to update cluster", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, updatedCluster)
	}

}
