package cluster

import (
	"config-service/dbhandler"
	"config-service/types"
	"config-service/utils/consts"
	"config-service/utils/log"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func postCluster(c *gin.Context) {
	clusters, err := dbhandler.MustGetDocContentFromContext[*types.Cluster](c)
	if err != nil {
		return
	}
	for i := range clusters {
		if clusters[i].Attributes == nil {
			clusters[i].Attributes = map[string]interface{}{}
		}
		clusters[i].Attributes[consts.ShortNameAttribute] = getUniqueShortName(clusters[i].Name, c)
	}
	dbhandler.PostDocHandler(c, clusters)
}

func putCluster(c *gin.Context) {
	docs, err := dbhandler.MustGetDocContentFromContext[*types.Cluster](c)
	if err != nil {
		return
	}
	reqCluster := docs[0]
	//only attributes can be updated - so check if there are any attributes
	if len(reqCluster.Attributes) == 0 {
		dbhandler.ResponseBadRequest(c, "cluster attributes are required")
		return
	}
	// if request attributes do not include alias add it from the old cluster
	if _, ok := reqCluster.Attributes[consts.ShortNameAttribute]; !ok {
		if oldCluster, err := dbhandler.GetDocByGUID[types.Cluster](c, reqCluster.GUID); err != nil {
			dbhandler.ResponseInternalServerError(c, "failed to read cluster", err)
			return
		} else if oldCluster == nil {
			dbhandler.ResponseDocumentNotFound(c)
			return
		} else {
			reqCluster.Attributes[consts.ShortNameAttribute] = oldCluster.Attributes[consts.ShortNameAttribute]
		}
	}
	//update only the attributes field
	update := dbhandler.GetUpdateFieldValueCommand(reqCluster.Attributes, consts.AttributesField)
	log.LogNTrace(fmt.Sprintf("post cluster %s - updating cluster", reqCluster.GUID), c)
	if oldAndUpdated, err := dbhandler.UpdateDocument[types.Cluster](c, reqCluster.GUID, update); err != nil {
		dbhandler.ResponseInternalServerError(c, "failed to read cluster", err)
		return
	} else if oldAndUpdated == nil {
		dbhandler.ResponseDocumentNotFound(c)
		return
	} else {
		c.JSON(http.StatusOK, oldAndUpdated)
	}
}
