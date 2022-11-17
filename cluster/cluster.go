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
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else {
		c.JSON(http.StatusOK, clusters)
	}
}

func getCluster(c *gin.Context) {
	guid := c.Param(utils.GUID_FIELD)
	if guid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster guid is required"})
		return
	}

	if cluster, err := mongo.GetDocByGUID(c, guid, &types.Cluster{}); err != nil {
		utils.LogNTraceError("failed to read cluster", err, c)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else {
		c.JSON(http.StatusOK, cluster)
	}
}

func postCluster(c *gin.Context) {
	reqCluster := types.Cluster{}
	if err := c.ShouldBindJSON(&reqCluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if reqCluster.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	nameFilter := mongo.NewFilterBuilder().
		WithNotDeleteForCustomer(c).
		WithValue("name", reqCluster.Name).
		Build()
	if count, err := mongo.GetReadCollection(utils.CLUSTERS).CountDocuments(c.Request.Context(), nameFilter); err != nil {
		utils.LogNTraceError("failed to count cluster with name", err, c)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	} else if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster name already exists"})
		return
	}
	clusterDoc := mongo.NewClusterDocument(reqCluster)
	clusterDoc.Customers = append(clusterDoc.Customers, c.GetString(utils.CUSTOMER_GUID))

	clusterDoc.Attributes[utils.ACRONYM_ATTRIBUTE] = getUniqueAcronym(clusterDoc.Name, c)

	if _, err := mongo.GetWriteCollection(utils.CLUSTERS).InsertOne(c.Request.Context(), clusterDoc); err != nil {
		utils.LogNTraceError("failed to create cluster", err, c)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, "created")
}

func putCluster(c *gin.Context) {
	reqCluster := types.Cluster{}
	if err := c.ShouldBindJSON(&reqCluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if guid := c.Param(utils.GUID_FIELD); guid != "" {
		reqCluster.GUID = guid
	}
	if reqCluster.GUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cluster guid is required"})
		return
	}
	if !mongo.DocExists(c, reqCluster.GUID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "cluster not found"})
		return
	}
	clusterDoc := mongo.NewClusterDocument(reqCluster)
	clusterDoc.ClearReadOnlyFields()
	if _, err := mongo.GetWriteCollection(utils.CLUSTERS).UpdateOne(c.Request.Context(),
		mongo.NewFilterBuilder().
			WithGUID(reqCluster.GUID).
			Build(),
		clusterDoc); err != nil {
		utils.LogNTraceError("failed to update cluster", err, c)
		c.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to update cluster"))
		return
	}
	c.JSON(http.StatusOK, "updated")
}

func getUniqueAcronym(name string, c *gin.Context) string {
	existingAcronyms := getAllAcronyms(c)
	return utils.MustGenerateAcronym(name, 5, existingAcronyms)
}

func getAllAcronyms(c *gin.Context) []string {
	if clusters, err := mongo.GetAllForCustomerWithProjection(c, []types.Cluster{}, mongo.NewProjectionBuilder().
		ExcludeID().
		Include(utils.ACRONYM_FIELD).
		Build()); err != nil {
		utils.LogNTraceError("failed to read clusters", err, c)
		return nil
	} else {
		var acrons []string
		for _, doc := range clusters {
			if doc.Attributes[utils.ACRONYM_ATTRIBUTE] != nil {
				acrons = append(acrons, doc.Attributes[utils.ACRONYM_ATTRIBUTE].(string))
			}
		}
		return acrons
	}
}
