package dbhandler

import (
	"config-service/mongo"
	"config-service/types"
	"config-service/utils/consts"
	"config-service/utils/log"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.mongodb.org/mongo-driver/bson"
	"k8s.io/utils/strings/slices"
)

/////////////////////////////////////////gin handlers/////////////////////////////////////////

// ////////////////////////////////////////GET///////////////////////////////////////////////

// HandleGetDocWithGUIDInPath - get document of type T by id in path
func HandleGetDocWithGUIDInPath[T types.DocContent](c *gin.Context) {
	guid := c.Param(consts.GUIDField)
	if guid == "" {
		ResponseMissingGUID(c)
		return
	}
	if doc, err := GetDocByGUID[T](c, guid); err != nil {
		ResponseInternalServerError(c, "failed to read document", err)
		return
	} else if doc == nil {
		ResponseDocumentNotFound(c)
		return
	} else {
		c.JSON(http.StatusOK, doc)
	}
}

// HandleGetListByNameOrAll - chains HandleGetNamesList->HandleGetByName-> HandleGetAll
func HandleGetByQueryOrAll[T types.DocContent](nameParam string, paramConf *scopeParamsConfig, listGlobals bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !GetNamesListHandler[T](c, listGlobals) &&
			!GetByNameParamHandler[T](c, nameParam) &&
			!GetByScopeParamsHandler[T](c, paramConf) {
			HandleGetAll[T](c)
		}
	}
}

// HandleGetAll - get all customer's documents of type T for collection in context
func HandleGetAll[T types.DocContent](c *gin.Context) {
	if docs, err := GetAllForCustomer[T](c, false); err != nil {
		ResponseInternalServerError(c, "failed to read all documents for customer", err)
		return
	} else {
		c.JSON(http.StatusOK, docs)
	}
}

// HandleGetAll - get all global and customer's documents of type T for collection in context
func HandleGetAllWithGlobals[T types.DocContent](c *gin.Context) {
	if docs, err := GetAllForCustomer[T](c, true); err != nil {
		ResponseInternalServerError(c, "failed to read all documents for customer", err)
		return
	} else {
		c.JSON(http.StatusOK, docs)
	}
}

// GetNamesList check for "list" query param and return list of names, returns false if not served by this handler
func GetNamesListHandler[T types.DocContent](c *gin.Context, includeGlobals bool) bool {
	if _, list := c.GetQuery(consts.ListParam); list {
		namesProjection := NewProjectionBuilder().Include(consts.NameField).ExcludeID().Get()
		if docNames, err := GetAllForCustomerWithProjection[T](c, namesProjection, includeGlobals); err != nil {
			ResponseInternalServerError(c, "failed to read documents", err)
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
		//get document by name
		if doc, err := GetDocByName[T](c, name); err != nil {
			ResponseInternalServerError(c, "failed to read document", err)
			return true
		} else if doc == nil {
			ResponseDocumentNotFound(c)
			return true
		} else {
			c.JSON(http.StatusOK, doc)
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
		if len(values) == 0 {
			continue
		}
		if len(keys) < 2 {
			keys = []string{conf.defaultContext, keys[0]}
		} else if len(keys) > 2 {
			keys = []string{keys[0], strings.Join(keys[1:], ".")}
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
		queryConfig, ok := conf.params2Query[field]
		if !ok {
			continue
		} else if queryConfig.isArray {
			if queryConfig.pathInArray != "" {
				key = queryConfig.pathInArray + "." + key

			}
		} else {
			key = queryConfig.fieldName + "." + key
		}
		//get the field filter builder
		filterBuilder := getFilterBuilder(queryConfig.fieldName)
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
		filterBuilder.WrapDupKeysWithOr()
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
		ResponseInternalServerError(c, "failed to read documents", err)
		return true
	} else {
		log.LogNTrace(fmt.Sprintf("scope query found %d documents", len(docs)), c)
		c.JSON(http.StatusOK, docs)
		return true
	}
}

// ////////////////////////////////////////POST///////////////////////////////////////////////
// HandlePostDocWithValidation - chains validation and post document handlers
func HandlePostDocWithValidation[T types.DocContent](validators ...Validator[T]) []gin.HandlerFunc {
	return []gin.HandlerFunc{HandlePostValidation(validators...), HandlePostDocFromContext[T]}
}

// HandlePostDocWithUniqueNameValidation - shortcut for HandlePostDocWithValidation(ValidateUniqueValues(NameKeyGetter[T]))
func HandlePostDocWithUniqueNameValidation[T types.DocContent]() []gin.HandlerFunc {
	return []gin.HandlerFunc{HandlePostValidation(ValidateUniqueValues(NameKeyGetter[T])), HandlePostDocFromContext[T]}
}

// HandlePutValidation validate post request and if valid sets one or many DocContents in context for next handler, otherwise abort request
func HandlePostValidation[T types.DocContent](validators ...Validator[T]) func(c *gin.Context) {
	return func(c *gin.Context) {
		var doc T
		var docs []T
		if err := c.ShouldBindBodyWith(&doc, binding.JSON); err != nil || doc == nil {
			//check if bulk request
			if err := c.ShouldBindBodyWith(&docs, binding.JSON); err != nil || docs == nil {
				ResponseFailedToBindJson(c, err)
				return
			}
		} else {
			//single request, append to slice
			docs = append(docs, doc)
		}

		//validate
		if len(docs) == 0 {
			ResponseBadRequest(c, "no documents in request")
			return
		}

		for _, validator := range validators {
			var ok bool
			if docs, ok = validator(c, docs); !ok {
				return
			}
		}
		c.Set(consts.DocContentKey, docs)
		c.Next()
	}
}

// HandlePostDocFromContext - handles creation of document(s) of type T
func HandlePostDocFromContext[T types.DocContent](c *gin.Context) {
	docs, err := MustGetDocContentFromContext[T](c)
	if err != nil {
		return
	}
	PostDocHandler(c, docs)
}

// PostDoc - helper to put document(s) of type T, custom handler should use this function to do the final POST handling
func PostDocHandler[T types.DocContent](c *gin.Context, docs []T) {
	collection, customerGUID, err := readContext(c)
	if err != nil {
		ResponseInternalServerError(c, "failed to read collection and customer guid from context", err)
		return
	}
	dbDocs := []interface{}{}
	for i := range docs {
		dbDocs = append(dbDocs, types.NewDocument(docs[i], customerGUID))
	}

	if len(dbDocs) == 1 {
		if _, err := mongo.GetWriteCollection(collection).InsertOne(c.Request.Context(), dbDocs[0]); err != nil {
			ResponseInternalServerError(c, "failed to create document", err)
			return
		} else {
			c.JSON(http.StatusCreated, docs[0])
		}
	} else {
		if _, err := mongo.GetWriteCollection(collection).InsertMany(c.Request.Context(), dbDocs); err != nil {
			ResponseInternalServerError(c, "failed to create document", err)
			return
		} else {
			c.JSON(http.StatusOK, docs)
		}
	}
}

// ////////////////////////////////////////PUT///////////////////////////////////////////////

// HandlePutDocWithValidation - chains validation and put document handlers
func HandlePutDocWithValidation[T types.DocContent](validators ...Validator[T]) []gin.HandlerFunc {
	return []gin.HandlerFunc{HandlePutValidation(validators...), HandlePutDocFromContext[T]}
}

// HandlePutDocWithGUIDValidation - shortcut for HandlePutDocWithValidation(ValidateGUIDExistence[T])
func HandlePutDocWithGUIDValidation[T types.DocContent]() []gin.HandlerFunc {
	return []gin.HandlerFunc{HandlePutValidation(ValidateGUIDExistence[T]), HandlePutDocFromContext[T]}
}

// HandlePutValidation validate put request and if valid set DocContent in context for next handler, otherwise abort request
func HandlePutValidation[T types.DocContent](validators ...Validator[T]) func(c *gin.Context) {
	return func(c *gin.Context) {
		var doc T
		if err := c.ShouldBindJSON(&doc); err != nil {
			ResponseFailedToBindJson(c, err)
			return
		}
		//validate
		for _, validator := range validators {
			if docs, ok := validator(c, []T{doc}); !ok {
				return
			} else {
				doc = docs[0]
			}
		}
		c.Set(consts.DocContentKey, doc)
		c.Next()
	}
}

// HandlePutDocFromContext - handles updates a document of type T
func HandlePutDocFromContext[T types.DocContent](c *gin.Context) {
	docs, err := MustGetDocContentFromContext[T](c)
	if err != nil {
		return
	}
	PutDocHandler(c, docs[0])
}

// PutDoc - helper to put document of type T, custom handler should use this function to do the final PUT handling
func PutDocHandler[T types.DocContent](c *gin.Context, doc T) {
	update, err := GetUpdateDocCommand(doc, doc.GetReadOnlyFields()...)
	if err != nil {
		ResponseInternalServerError(c, "failed to generate update command", err)
		return
	}
	if res, err := UpdateDocument[T](c, doc.GetGUID(), update); err != nil {
		ResponseInternalServerError(c, "failed to update document", err)
	} else if res == nil {
		ResponseDocumentNotFound(c)
		return
	} else {
		c.JSON(http.StatusOK, res)
	}
}

// ////////////////////////////////////////DELETE///////////////////////////////////////////////

// HandleDeleteDoc  - delete document by id in path
func HandleDeleteDoc[T types.DocContent](c *gin.Context) {
	guid := c.Param(consts.GUIDField)
	if guid == "" {
		ResponseMissingGUID(c)
		return
	}
	DeleteDocByGUIDHandler[T](c, guid)
}

// HandleDeleteDocByName  - delete document(s) by name in path
func HandleDeleteDocByName[T types.DocContent](nameParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		names, ok := c.GetQueryArray(nameParam)
		names = slices.Filter([]string{}, names, func(s string) bool {
			return s != ""
		})
		if !ok || len(names) == 0 {
			ResponseMissingName(c)
			return
		}
		if len(names) == 1 {
			DeleteDocByNameHandler[T](c, names[0])
		} else {
			BulkDeleteDocByNameHandler[T](c, names)
		}

	}
}

func BulkDeleteDocByNameHandler[T types.DocContent](c *gin.Context, names []string) {
	collection, err := readCollection(c)
	if err != nil {
		ResponseInternalServerError(c, "failed to read collection from context", err)
		return
	}
	filter := NewFilterBuilder().WithIn("name", names).WithNotDeleteForCustomer(c)
	if res, err := mongo.GetWriteCollection(collection).DeleteMany(c.Request.Context(), filter.Get()); err != nil {
		ResponseInternalServerError(c, "failed to delete document", err)
	} else if res.DeletedCount == 0 {
		ResponseDocumentNotFound(c)
	} else {
		c.JSON(http.StatusOK, res)
	}
}

func DeleteDocByGUIDHandler[T types.DocContent](c *gin.Context, guid string) {
	collection, err := readCollection(c)
	if err != nil {
		ResponseInternalServerError(c, "failed to read collection from context", err)
		return
	}
	toBeDeleted, err := GetDocByGUID[T](c, guid)
	if err != nil {
		ResponseInternalServerError(c, "failed to read document", err)
		return
	} else if toBeDeleted == nil {
		ResponseDocumentNotFound(c)
		return
	}
	if res, err := mongo.GetWriteCollection(collection).DeleteOne(c.Request.Context(), bson.M{consts.IdField: guid}); err != nil {
		ResponseInternalServerError(c, "failed to delete document", err)
		return
	} else if res.DeletedCount == 0 {
		ResponseDocumentNotFound(c)
		return
	}
	c.JSON(http.StatusOK, toBeDeleted)
}

func DeleteDocByNameHandler[T types.DocContent](c *gin.Context, name string) {
	collection, err := readCollection(c)
	if err != nil {
		ResponseInternalServerError(c, "failed to read collection from context", err)
		return
	}
	toBeDeleted, err := GetDocByName[T](c, name)
	if err != nil {
		ResponseInternalServerError(c, "failed to read document", err)
		return
	} else if toBeDeleted == nil {
		ResponseDocumentNotFound(c)
		return
	}

	if res, err := mongo.GetWriteCollection(collection).DeleteOne(c.Request.Context(), bson.M{consts.IdField: (*toBeDeleted).GetGUID()}); err != nil {
		ResponseInternalServerError(c, "failed to delete document", err)
		return
	} else if res.DeletedCount == 0 {
		ResponseDocumentNotFound(c)
		return
	}
	c.JSON(http.StatusOK, toBeDeleted)
}
