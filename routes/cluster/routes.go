package cluster

import (
	"config-service/db"
	"config-service/handlers"
	"config-service/types"
	"config-service/utils/consts"
	"config-service/utils/log"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {

	handlers.AddRoutes(g,
		handlers.WithPath[*types.Cluster](consts.ClusterPath),
		handlers.WithDBCollection[*types.Cluster](consts.ClustersCollection),
		handlers.WithValidatePostUniqueName[*types.Cluster](true),
		handlers.WithValidatePutGUID[*types.Cluster](true),
		handlers.WithDeleteByName[*types.Cluster](false),
		handlers.WithPostValidators(validatePostClusterShortNames),
		handlers.WithPutValidators(validatePutClusterShortNames),
	)
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
			handlers.ResponseBadRequest(c, "cluster attributes are required")
			return nil, false
		}
		// if request attributes do not include alias add it from the old cluster
		if _, ok := clusters[i].Attributes[consts.ShortNameAttribute]; !ok {
			if oldCluster, err := db.GetDocByGUID[types.Cluster](c, clusters[i].GUID); err != nil {
				handlers.ResponseInternalServerError(c, "failed to read cluster", err)
				return nil, false
			} else if oldCluster == nil {
				handlers.ResponseDocumentNotFound(c)
				return nil, false
			} else {
				clusters[i].Attributes[consts.ShortNameAttribute] = oldCluster.Attributes[consts.ShortNameAttribute]
			}
		}
	}
	return clusters, true
}
