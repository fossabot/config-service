package dbhandler

import (
	"fmt"
	"kubescape-config-service/mongo"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"
	"kubescape-config-service/utils/log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)



//////////////////////////////////Sugar functions for mongo using values in gin context /////////////////////////////////////////
/////////////////////////////////all methods are expecting collection and customerGUID from context/////////////////////////////

// GetAllForCustomer returns all docs for customer
func GetAllForCustomer[T any](c *gin.Context) ([]T, error) {
	return GetAllForCustomerWithProjection[T](c, nil)
}

// GetAllForCustomerWithProjection returns all docs for customer with projection
func GetAllForCustomerWithProjection[T any](c *gin.Context, projection bson.D) ([]T, error) {
	collection, _, err := readContext(c)
	result := []T{}
	if err != nil {
		return nil, err
	}
	filter := NewFilterBuilder().WithNotDeleteForCustomer(c).Get()
	findOpts := options.Find().SetNoCursorTimeout(true)
	if projection != nil {
		findOpts.SetProjection(projection)
	}
	if cur, err := mongo.GetReadCollection(collection).
		Find(c.Request.Context(), filter, findOpts); err != nil {
		return nil, err
	} else {

		if err := cur.All(c.Request.Context(), &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func FindForCustomer[T any](c *gin.Context, filterBuilder *FilterBuilder, projection bson.D) ([]T, error) {
	collection, _, err := readContext(c)
	result := []T{}
	if err != nil {
		return nil, err
	}
	if filterBuilder == nil {
		filterBuilder = NewFilterBuilder()
	}
	filter := filterBuilder.WithNotDeleteForCustomer(c).Get()
	findOpts := options.Find().SetNoCursorTimeout(true)
	if projection != nil {
		findOpts.SetProjection(projection)
	}
	if cur, err := mongo.GetReadCollection(collection).
		Find(c.Request.Context(), filter, findOpts); err != nil {
		return nil, err
	} else {

		if err := cur.All(c.Request.Context(), &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// UpdateDocument updates document by GUID and update command
func UpdateDocument[T any](c *gin.Context, id string, update bson.D) ([]T, error) {

	collection, _, err := readContext(c)
	if err != nil {
		return nil, err
	}

	var oldDoc T
	if err := mongo.GetReadCollection(collection).
		FindOne(c.Request.Context(),
			NewFilterBuilder().
				WithNotDeleteForCustomer(c).
				WithID(id).
				Get()).
		Decode(&oldDoc); err != nil {
		log.LogNTraceError("failed to get document by id", err, c)
		return nil, err
	}
	var newDoc T
	filter := NewFilterBuilder().WithNotDeleteForCustomer(c).WithID(id).Get()
	if err := mongo.GetWriteCollection(collection).FindOneAndUpdate(c.Request.Context(), filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).
		Decode(&newDoc); err != nil {
		return nil, err
	}
	return []T{oldDoc, newDoc}, nil
}

// DocExist returns true if at least one document with given filter exists
func DocExist(c *gin.Context, f bson.D) (bool, error) {
	collection, _, err := readContext(c)
	if err != nil {
		return false, err
	}
	filter := NewFilterBuilder().
		WithNotDeleteForCustomer(c).
		WithFilter(f).
		Get()
	n, err := mongo.GetReadCollection(collection).CountDocuments(c.Request.Context(), filter, options.Count().SetLimit(1))
	return n > 0, err
}

// DocWithNameExist returns true if at least one document with given name exists
func DocWithNameExist(c *gin.Context, name string) (bool, error) {
	return DocExist(c,
		NewFilterBuilder().
			WithName(name).
			Get())
}

// GetDocByGUID returns document by GUID
func GetDocByGUID[T any](c *gin.Context, guid string) (*T, error) {
	collection, _, err := readContext(c)
	if err != nil {
		return nil, err
	}
	var result T
	if err := mongo.GetReadCollection(collection).
		FindOne(c.Request.Context(),
			NewFilterBuilder().
				WithNotDeleteForCustomer(c).
				WithGUID(guid).
				Get()).
		Decode(&result); err != nil {
		log.LogNTraceError("failed to get document by id", err, c)
		return nil, err
	}
	return &result, nil
}

// GetDocByName returns document by name
func GetDocByName[T any](c *gin.Context, name string) (*T, error) {
	collection, _, err := readContext(c)
	if err != nil {
		return nil, err
	}
	var result T
	if err := mongo.GetReadCollection(collection).
		FindOne(c.Request.Context(),
			NewFilterBuilder().
				WithNotDeleteForCustomer(c).
				WithName(name).
				Get()).
		Decode(&result); err != nil {
		log.LogNTraceError("failed to get document by id", err, c)
		return nil, err
	}
	return &result, nil
}

// CountDocs counts documents that match the filter
func CountDocs(c *gin.Context, f bson.D) (int64, error) {
	collection, _, err := readContext(c)
	if err != nil {
		return 0, err
	}
	filter := NewFilterBuilder().
		WithNotDeleteForCustomer(c).
		WithFilter(f).
		Get()
	return mongo.GetReadCollection(collection).CountDocuments(c.Request.Context(), filter)
}

// helpers

func MustGetDocContentFromContext[T types.DocContent](c *gin.Context) (T, error) {
	var doc T
	if iData, ok := c.Get(consts.DOC_CONTENT_KEY); ok {
		doc = iData.(T)
	} else {
		err := fmt.Errorf("failed to get doc content from context")
		log.LogNTraceError(err.Error(), err, c)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, err
	}
	return doc, nil
}

// readContext reads collection and customerGUID from context
func readContext(c *gin.Context) (collection, customerGUID string, err error) {
	collection, errCollection := readCollection(c)
	if errCollection != nil {
		err = multierror.Append(err, errCollection)
	}
	customerGUID, errGuid := readCustomerGUID(c)
	if errGuid != nil {
		err = multierror.Append(err, errGuid)
	}
	return collection, customerGUID, err
}

func readCustomerGUID(c *gin.Context) (customerGUID string, err error) {
	customerGUID = c.GetString(consts.CUSTOMER_GUID)
	if customerGUID == "" {
		err = fmt.Errorf("customerGUID is not in context")
	}
	return customerGUID, err
}

func readCollection(c *gin.Context) (collection string, err error) {
	collection = c.GetString(consts.COLLECTION)
	if collection == "" {
		err = fmt.Errorf("collection is not in context")
	}
	return collection, err
}
