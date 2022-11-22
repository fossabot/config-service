package dbhandler

import (
	"fmt"
	"kubescape-config-service/mongo"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"
	"kubescape-config-service/utils/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

/////////////////////////////////////////gin handlers/////////////////////////////////////////

// HandleDeleteDoc  - delete document by id in path for collection in context
func HandleDeleteDoc(c *gin.Context) {
	collection, _, err := readContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	guid := c.Param(consts.GUID_FIELD)
	if guid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "guid is required"})
		return
	}

	if res, err := mongo.GetWriteCollection(collection).DeleteOne(c.Request.Context(), bson.M{consts.ID_FIELD: guid}); err != nil {
		msg := fmt.Sprintf("failed to delete document GUID: %s  Collection: %s", guid, collection)
		log.LogNTraceError(msg, err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, fmt.Sprintf("document with id %s does not exist", guid))
		return
	}
	c.JSON(http.StatusOK, "deleted")
}

// HandleGetDocWithGUIDInPath - get document of type T by id in path for collection in context
func HandleGetDocWithGUIDInPath[T types.DocContent](c *gin.Context) {
	guid := c.Param(consts.GUID_FIELD)
	if guid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "guid is required"})
		return
	}
	if policy, err := GetDocByGUID[T](c, guid); err != nil {
		log.LogNTraceError("failed to read document", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, policy)
	}
}

// HandleGetAll - get all documents of type T for collection in context
func HandleGetAll[T types.DocContent](c *gin.Context) {
	if docs, err := GetAllForCustomer[T](c); err != nil {
		log.LogNTraceError("failed to read all documents for customer", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, docs)
	}
}

// HandlePostDocWithValidation - chains validation and post document handlers
func HandlePostDocWithValidation[T types.DocContent]() []gin.HandlerFunc {
	return []gin.HandlerFunc{HandlePostValidation[T], HandlePostDocFromContext[T]}
}

// HandlePutDocWithValidation - chains validation and put document handlers
func HandlePutDocWithValidation[T types.DocContent]() []gin.HandlerFunc {
	return []gin.HandlerFunc{HandlePutValidation[T], HandlePutDocFromContext[T]}
}

// HandlePostDocFromContext - post document of type T from context
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

// HandlePutDocFromContext - put document of type T from context
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

// HandlePostValidation validate post request and if valid set DocContent in context for next handler, otherwise abort request

func HandlePostValidation[T types.DocContent](c *gin.Context) {
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
			WithName(doc.GetName()).
			Get()); err != nil {
		log.LogNTraceError("HandlePostValidation: failed to check if document with same name exist", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if exist {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("document with name %s already exists", doc.GetName())})
		return
	}
	c.Set("docData", doc)
	c.Next()
}

// HandlePutValidation validate put request and if valid set DocContent in context for next handler, otherwise abort request
func HandlePutValidation[T types.DocContent](c *gin.Context) {
	var doc T
	if err := c.BindJSON(&doc); err != nil {
		return
	}
	if guid := c.Param(consts.GUID_FIELD); guid != "" {
		doc.SetGUID(guid)
	}
	if doc.GetGUID() == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cluster guid is required"})
		return
	}
	c.Set("docData", doc)
	c.Next()
}

// PostDoc - helper to post document of type T an be used by custom handlers
func PostDoc[T types.DocContent](c *gin.Context, doc T) {
	collection, customerGUID, err := readContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dbDoc := NewDocument(doc, customerGUID)
	if result, err := mongo.GetWriteCollection(collection).InsertOne(c.Request.Context(), dbDoc); err != nil {
		log.LogNTraceError("failed to create document", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"GUID": result.InsertedID})
	}
}

// PutDoc - helper to put document of type T an be used by custom handlers
func PutDoc[T types.DocContent](c *gin.Context, doc T) {
	update, err := GetUpdateDocCommand(doc, doc.GetReadOnlyFields()...)
	if err != nil {
		log.LogNTraceError("failed to create update command", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if res, err := UpdateDocument[T](c, doc.GetGUID(), update); err != nil {
		log.LogNTraceError("failed to update document", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, res)
	}
}
