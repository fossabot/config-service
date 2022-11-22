package dbhandler

import (
	"fmt"
	"kubescape-config-service/mongo"
	"kubescape-config-service/types"
	"kubescape-config-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

/////////////////////////////////////////gin handlers/////////////////////////////////////////

// HandleDeleteDoc gin handler for delete document by id in collection in context
func HandleDeleteDoc(c *gin.Context) {
	collection, _, err := readContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	guid := c.Param(utils.GUID_FIELD)
	if guid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "guid is required"})
		return
	}

	if res, err := mongo.GetWriteCollection(collection).DeleteOne(c.Request.Context(), bson.M{utils.ID_FIELD: guid}); err != nil {
		msg := fmt.Sprintf("failed to delete document GUID: %s  Collection: %s", guid, collection)
		utils.LogNTraceError(msg, err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, fmt.Sprintf("document with id %s does not exist", guid))
	}
	c.JSON(http.StatusOK, "deleted")
}

func HandleGetDocWithGUIDInPath[T types.DocContent](c *gin.Context) {
	guid := c.Param(utils.GUID_FIELD)
	if guid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "guid is required"})
		return
	}
	var doc T
	if policy, err := GetDocByGUID(c, guid, &doc); err != nil {
		utils.LogNTraceError("failed to read document", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, policy)
	}
}

func HandleGetAll[T types.DocContent](c *gin.Context) {
	if docs, err := GetAllForCustomer(c, []T{}); err != nil {
		utils.LogNTraceError("failed to read all documents for customer", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, docs)
	}
}
func HandlePostDocWithValidation[T types.DocContent]() []gin.HandlerFunc {
	return []gin.HandlerFunc{PostValidation[T], HandlePostDocFromContext[T]}
}

func HandlePutDocWithValidation[T types.DocContent]() []gin.HandlerFunc {
	return []gin.HandlerFunc{PutValidation[T], HandlePutDocFromContext[T]}
}

func HandlePostDocFromContext[T types.DocContent](c *gin.Context) {
	var doc T
	if iData, ok := c.Get("docData"); ok {
		doc = iData.(T)
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "docData is required"})
		return
	}
	PostDoc(c, doc)
}

func HandlePutDocFromContext[T types.DocContent](c *gin.Context) {
	var doc T
	if iData, ok := c.Get("docData"); ok {
		doc = iData.(T)
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "docData is required"})
		return
	}
	PutDoc(c, doc)
}

func PostDoc[T types.DocContent](c *gin.Context, doc T) {
	collection, customerGUID, err := readContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dbDoc := NewDocument(doc, customerGUID)
	if result, err := mongo.GetWriteCollection(collection).InsertOne(c.Request.Context(), dbDoc); err != nil {
		utils.LogNTraceError("failed to create document", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"GUID": result.InsertedID})
	}
}

func PutDoc[T types.DocContent](c *gin.Context, doc T) {
	update, err := GetUpdateDocCommand(doc, doc.GetReadOnlyFields()...)
	if err != nil {
		utils.LogNTraceError("failed to create update command", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var res T
	if res, err := UpdateDocument(c, doc.GetGUID(), update, &res); err != nil {
		utils.LogNTraceError("failed to update document", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, res)
	}
}

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
