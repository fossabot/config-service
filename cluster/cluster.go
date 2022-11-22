package cluster

import (
	"fmt"
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"
	"kubescape-config-service/utils/log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func postCluster(c *gin.Context) {
	var reqCluster *types.Cluster
	if iData, ok := c.Get("docData"); ok {
		reqCluster = iData.(*types.Cluster)
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "docData is required"})
		return
	}
	if reqCluster.Attributes == nil {
		reqCluster.Attributes = map[string]interface{}{}
	}
	reqCluster.Attributes[consts.SHORT_NAME_ATTRIBUTE] = getUniqueShortName(reqCluster.Name, c)
	dbhandler.PostDoc(c, reqCluster)
}

func putCluster(c *gin.Context) {
	var reqCluster *types.Cluster
	//get cluster body from context
	if iData, ok := c.Get("docData"); ok {
		reqCluster = iData.(*types.Cluster)
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "docData is required"})
		return
	}
	//only attributes can be updated - so check if there are any attributes
	if len(reqCluster.Attributes) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cluster attributes are required"})
		return
	}
	// if request attributes do not include alias add it from the old cluster
	if _, ok := reqCluster.Attributes[consts.SHORT_NAME_ATTRIBUTE]; !ok {
		if oldCluster, err := dbhandler.GetDocByGUID[types.Cluster](c, reqCluster.GUID); err != nil {
			log.LogNTraceError("failed to read cluster", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else {
			reqCluster.Attributes[consts.SHORT_NAME_ATTRIBUTE] = oldCluster.Attributes[consts.SHORT_NAME_ATTRIBUTE]
		}
	}
	//update only the attributes field
	update := dbhandler.GetUpdateFieldValueCommand(reqCluster.Attributes, consts.ATTRIBUTES_FIELD)
	log.LogNTrace(fmt.Sprintf("post cluster %s - updating cluster", reqCluster.GUID), c)
	if updatedCluster, err := dbhandler.UpdateDocument[types.Cluster](c, reqCluster.GUID, update); err != nil {
		log.LogNTraceError("failed to update cluster", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, updatedCluster)
	}
}
