package dbhandler

import (
	"fmt"
	"kubescape-config-service/mongo"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"
	"kubescape-config-service/utils/log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	mongoDB "go.mongodb.org/mongo-driver/mongo"
	"k8s.io/utils/strings/slices"
)

/////////////////////////////////////////gin handlers/////////////////////////////////////////

// ////////////////////////////////////////GET///////////////////////////////////////////////

// HandleGetDocWithGUIDInPath - get document of type T by id in path
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

// HandleGetListByNameOrAll - chains HandleGetNamesList->HandleGetByName-> HandleGetAll
func HandleGetByQueryOrAll[T types.DocContent](nameParam string, paramConf *scopeParamsConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !GetNamesListHandler[T](c) &&
			!GetByNameParamHandler[T](c, nameParam) &&
			!GetByScopeParamsHandler[T](c, paramConf) {
			HandleGetAll[T](c)
		}
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

// GetNamesList check for "list" query param and return list of names, returns false if not served by this handler
func GetNamesListHandler[T types.DocContent](c *gin.Context) bool {
	if _, list := c.GetQuery(consts.LIST_PARAM); list {
		namesProjection := NewProjectionBuilder().Include(consts.NAME_FIELD).ExcludeID().Get()
		if docNames, err := GetAllForCustomerWithProjection[T](c, namesProjection); err != nil {
			log.LogNTraceError("failed to read documents with name projection", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return true
		} else {
			var names []string
			for _, docContent := range docNames {
				names = append(names, docContent.GetName())
			}
			c.JSON(http.StatusOK, names)
			return true
		}
	}
	return false
}

// HandleGetNameList check for <nameParam> query param and return the element with this name, returns false if not served by this handler
func GetByNameParamHandler[T types.DocContent](c *gin.Context, nameParam string) bool {
	if nameParam == "" {
		return false
	}
	if name := c.Query(nameParam); name != "" {
		//get policy by name
		if policy, err := GetDocByName[T](c, name); err != nil {
			log.LogNTraceError("failed to read policy", err, c)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return true
		} else {
			c.JSON(http.StatusOK, policy)
			c.Done()
			return true
		}
	}
	return false
}

// GetByScopeParams parse scope params and return elements with this scope, returns false if not served by this handler
func GetByScopeParamsHandler[T types.DocContent](c *gin.Context, conf *scopeParamsConfig) bool {
	if conf == nil {
		return false // not served by this handler
	}

	//keep filter builder per field name
	filterBuilders := map[string]*FilterBuilder{}
	getFilterBuilder := func(paramName string) *FilterBuilder {
		if filterBuilder, ok := filterBuilders[paramName]; ok {
			return filterBuilder
		}
		filterBuilder := NewFilterBuilder()
		filterBuilders[paramName] = filterBuilder
		return filterBuilder
	}

	qParams := c.Request.URL.Query()
	for paramKey, vals := range qParams {
		keys := strings.Split(paramKey, ".")
		//clean whitespaces
		values := slices.Filter([]string{}, vals, func(s string) bool { return s != "" })
		if len(keys) != 2 || len(values) == 0 {
			err := fmt.Errorf("invalid query param %s %s", paramKey, strings.Join(values, ","))
			log.LogNTraceError("invalid query param", err, c)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return true
		}
		//escape in case of bad formatted query params
		for i := range values {
			if v, err := url.QueryUnescape(values[i]); err != nil {
				log.LogNTraceError("failed to unescape query param", err, c)
			} else {
				values[i] = v
			}
		}
		//calculate field name
		var field, key = keys[0], keys[1]
		if queryConfig, ok := conf.params2Query[field]; !ok {
			continue
		} else if queryConfig.isArray {
			if queryConfig.pathInArray != "" {
				key = queryConfig.pathInArray + "." + key

			}
		} else {
			key = queryConfig.fieldName + "." + key
		}
		//get the field filter builder
		filterBuilder := getFilterBuilder(field)
		//case of single value
		if len(values) == 1 {
			filterBuilder.WithValue(key, values[0])
		} else { //case of multiple values
			fb := NewFilterBuilder()
			for _, v := range values {
				fb.WithValue(key, v)
			}
			filterBuilder.WithFilter(fb.WarpOr().Get())
		}
	}
	//aggregate all filters
	allQueriesFilter := NewFilterBuilder()
	for key, filterBuilder := range filterBuilders {
		queryConfig := conf.params2Query[key]
		if queryConfig.isArray {
			filterBuilder.WarpElementMatch().WarpWithField(queryConfig.fieldName)
		}
		allQueriesFilter.WithFilter(filterBuilder.Get())
	}
	if len(allQueriesFilter.Get()) == 0 {
		return false //not served by this handler
	}
	log.LogNTrace(fmt.Sprintf("query params: %v search query %v", qParams, allQueriesFilter.Get()), c)
	if docs, err := FindForCustomer[T](c, allQueriesFilter, nil); err != nil {
		log.LogNTraceError("failed to read documents", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return true
	} else {
		log.LogNTrace(fmt.Sprintf("scope query found %d documents", len(docs)), c)
		c.JSON(http.StatusOK, docs)
		return true
	}
}

// ////////////////////////////////////////POST///////////////////////////////////////////////
// HandlePostDocWithValidation - chains validation and post document handlers
func HandlePostDocWithValidation[T types.DocContent]() []gin.HandlerFunc {
	return []gin.HandlerFunc{HandlePostValidation[T], HandlePostDocFromContext[T]}
}

// HandlePutValidation validate post request and if valid set DocContent in context for next handler, otherwise abort request
func HandlePostValidation[T types.DocContent](c *gin.Context) {
	var doc T
	if err := c.ShouldBindJSON(&doc); err != nil {
		log.LogNTraceError("failed to bind json", err, c)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
	c.Set(consts.DOC_CONTENT_KEY, doc)
	c.Next()
}

// HandlePutDocFromContext - handles create a document of type T
func HandlePostDocFromContext[T types.DocContent](c *gin.Context) {
	doc, err := MustGetDocContentFromContext[T](c)
	if err != nil {
		return
	}
	PostDocHandler(c, doc)
}

// PostDoc - helper to put document of type T, custom handler should use this function to do the final POST handling
func PostDocHandler[T types.DocContent](c *gin.Context, doc T) {
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

// ////////////////////////////////////////PUT///////////////////////////////////////////////

// HandlePutDocWithValidation - chains validation and put document handlers
func HandlePutDocWithValidation[T types.DocContent]() []gin.HandlerFunc {
	return []gin.HandlerFunc{HandlePutValidation[T], HandlePutDocFromContext[T]}
}

// HandlePutValidation validate put request and if valid set DocContent in context for next handler, otherwise abort request
func HandlePutValidation[T types.DocContent](c *gin.Context) {
	var doc T
	if err := c.ShouldBindJSON(&doc); err != nil {
		log.LogNTraceError("failed to bind json", err, c)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if guid := c.Param(consts.GUID_FIELD); guid != "" {
		doc.SetGUID(guid)
	}
	if doc.GetGUID() == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cluster guid is required"})
		return
	}
	c.Set(consts.DOC_CONTENT_KEY, doc)
	c.Next()
}

// HandlePutDocFromContext - handles updates a document of type T
func HandlePutDocFromContext[T types.DocContent](c *gin.Context) {
	doc, err := MustGetDocContentFromContext[T](c)
	if err != nil {
		return
	}
	PutDocHandler(c, doc)
}

// PutDoc - helper to put document of type T, custom handler should use this function to do the final PUT handling
func PutDocHandler[T types.DocContent](c *gin.Context, doc T) {
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

// ////////////////////////////////////////DELETE///////////////////////////////////////////////

// HandleDeleteDoc  - delete document by id in path for collection in context
func HandleDeleteDoc[T types.DocContent](c *gin.Context) {
	guid := c.Param(consts.GUID_FIELD)
	if guid == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "guid is required"})
		return
	}
	DeleteDocByGUIDHandler[T](c, guid)
}

// HandleDeleteDoc  - delete document by id in path for collection in context
func HandleDeleteDocByName[T types.DocContent](nameParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query(nameParam)
		if name == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "guid is required"})
			return
		}
		DeleteDocByNameHandler[T](c, name)
	}
}

func DeleteDocByGUIDHandler[T types.DocContent](c *gin.Context, guid string) {
	collection, _, err := readContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	toBeDeleted, err := GetDocByGUID[T](c, guid)
	if err != nil {
		if err == mongoDB.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("document with guid %s not found", guid)})
			return
		}
		log.LogNTraceError("failed to read document before delete", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	c.JSON(http.StatusOK, toBeDeleted)
}

func DeleteDocByNameHandler[T types.DocContent](c *gin.Context, name string) {
	collection, _, err := readContext(c)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	toBeDeleted, err := GetDocByName[T](c, name)
	if err != nil {
		if err == mongoDB.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("document with name %s not found", name)})
			return
		}
		log.LogNTraceError("failed to read document before delete", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filter := NewFilterBuilder().WithNotDeleteForCustomer(c).WithName(name)

	if res, err := mongo.GetWriteCollection(collection).DeleteOne(c.Request.Context(), filter); err != nil {
		msg := fmt.Sprintf("failed to delete document with name: %s  Collection: %s", name, collection)
		log.LogNTraceError(msg, err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, fmt.Sprintf("document with name %s does not exist", name))
		return
	}
	c.JSON(http.StatusOK, toBeDeleted)
}
