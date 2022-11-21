package dbhandler

import (
	"fmt"
	"kubescape-config-service/types"
	"kubescape-config-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

/////////////////////////////////////////gin middleware/////////////////////////////////////////

func PostValidation[T types.DocContent](c *gin.Context) {
	var doc T
	if err := c.BindJSON(&doc); err != nil {
		return
	}
	if doc.GetName() == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if exist, err := DocExist(c,
		NewFilterBuilder().
			WithValue("name", doc.GetName()).
			Get()); err != nil {
		utils.LogNTraceError("PostValidation: failed to check if document with same name exist", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("document with name %s already exists", doc.GetName())})
		return
	}
	c.Set("docData", doc)
	c.Next()
}

func PutValidation[T types.DocContent](c *gin.Context) {
	var doc T
	if err := c.BindJSON(&doc); err != nil {
		return
	}
	if guid := c.Param(utils.GUID_FIELD); guid != "" {
		doc.SetGUID(guid)
	}
	if doc.GetGUID() == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cluster guid is required"})
		return
	}
	c.Set("docData", doc)
	c.Next()
}
