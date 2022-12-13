package cluster

import (
	"config-service/dbhandler"
	"config-service/types"
	"config-service/utils/consts"
	"config-service/utils/log"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	cluster := g.Group(consts.ClusterPath)

	cluster.Use(dbhandler.DBContextMiddleware(consts.ClustersCollection))

	cluster.GET("", dbhandler.HandleGetAll[*types.Cluster])
	cluster.GET("/:"+consts.GUIDField, dbhandler.HandleGetDocWithGUIDInPath[*types.Cluster])
	cluster.POST("", dbhandler.HandlePostDocWithValidation(dbhandler.ValidateUniqueValues(dbhandler.NameKeyGetter[*types.Cluster]), validatePostClusterShortNames)...)
	cluster.PUT("", dbhandler.HandlePutDocWithValidation(dbhandler.ValidateGUIDExistence[*types.Cluster], validatePutClusterShortNames)...)
	cluster.PUT("/:"+consts.GUIDField, dbhandler.HandlePutDocWithValidation(dbhandler.ValidateGUIDExistence[*types.Cluster], validatePutClusterShortNames)...)
	cluster.DELETE("/:"+consts.GUIDField, dbhandler.HandleDeleteDoc[*types.Cluster])
}

func validatePostClusterShortNames(c *gin.Context, clusters []*types.Cluster) ([]*types.Cluster, bool) {
	defer log.LogNTraceEnterExit("validatePostClusterShortNames", c)()
	for i := range clusters {
		if clusters[i].Attributes == nil {
			clusters[i].Attributes = map[string]interface{}{}
		}
		clusters[i].Attributes[consts.ShortNameAttribute] = getUniqueShortName(clusters[i].Name, c)
	}
	return clusters, true
}

func validatePutClusterShortNames(c *gin.Context, clusters []*types.Cluster) ([]*types.Cluster, bool) {
	defer log.LogNTraceEnterExit("validatePutClusterShortNames", c)()
	for i := range clusters {
		if len(clusters[i].Attributes) == 0 {
			dbhandler.ResponseBadRequest(c, "cluster attributes are required")
			return nil, false
		}
		// if request attributes do not include alias add it from the old cluster
		if _, ok := clusters[i].Attributes[consts.ShortNameAttribute]; !ok {
			if oldCluster, err := dbhandler.GetDocByGUID[types.Cluster](c, clusters[i].GUID); err != nil {
				dbhandler.ResponseInternalServerError(c, "failed to read cluster", err)
				return nil, false
			} else if oldCluster == nil {
				dbhandler.ResponseDocumentNotFound(c)
				return nil, false
			} else {
				clusters[i].Attributes[consts.ShortNameAttribute] = oldCluster.Attributes[consts.ShortNameAttribute]
			}
		}
	}
	return clusters, true
}
