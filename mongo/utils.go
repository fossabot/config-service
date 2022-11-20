package mongo

import (
	"fmt"
	"kubescape-config-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/////////////////////////////////////////gin handlers/////////////////////////////////////////

//HandleDeleteDoc gin handler for delete document by id in collection in context
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

	if res, err := GetWriteCollection(collection).DeleteOne(c.Request.Context(), bson.M{utils.ID_FIELD: guid}); err != nil {
		msg := fmt.Sprintf("failed to delete document GUID: %s  Collection: %s", guid, collection)
		utils.LogNTraceError(msg, err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, fmt.Sprintf("document with id %s does not exist", guid))
	}
	c.JSON(http.StatusOK, "deleted")
}

//////////////////////////////////Sugar functions for mongo using values in gin context ///////////////////////////////////////////

//GetAllForCustomer returns all not delete docs for customer from customerGUID and collection in context
func GetAllForCustomer[T any](c *gin.Context, result []T) ([]T, error) {
	return GetAllForCustomerWithProjection(c, result, nil)
}

//GetAllForCustomerWithProjection returns all not delete docs for customer from customerGUID and collection in context
func GetAllForCustomerWithProjection[T any](c *gin.Context, result []T, projection bson.D) ([]T, error) {
	collection, _, err := readContext(c)
	if err != nil {
		return nil, err
	}
	filter := NewFilterBuilder().WithNotDeleteForCustomer(c).Get()
	findOpts := options.Find().SetNoCursorTimeout(true)
	if projection != nil {
		findOpts.SetProjection(projection)
	}
	if cur, err := GetReadCollection(collection).
		Find(c.Request.Context(), filter, findOpts); err != nil {
		return nil, err
	} else {
		if err := cur.All(c.Request.Context(), &result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func UpdateDocument[T any](c *gin.Context, id string, update bson.D, result *T) (*T, error) {
	collection, _, err := readContext(c)
	if err != nil {
		return nil, err
	}
	filter := NewFilterBuilder().WithNotDeleteForCustomer(c).WithID(id).Get()
	if err := GetWriteCollection(collection).FindOneAndUpdate(c.Request.Context(), filter, update,
		options.FindOneAndUpdate().SetReturnDocument(options.After)).
		Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}

//DocExist returns true if at least one document with given filter exists for customer & collection in context
func DocExist(c *gin.Context, f bson.D) (bool, error) {
	collection, _, err := readContext(c)
	if err != nil {
		return false, err
	}
	filter := NewFilterBuilder().
		WithNotDeleteForCustomer(c).
		WithFilter(f).
		Get()
	n, err := GetReadCollection(collection).CountDocuments(c.Request.Context(), filter, options.Count().SetLimit(1))
	return n > 0, err
}

//GetDocByGUID returns document by GUID for customer in context from collection in context
func GetDocByGUID[T any](c *gin.Context, guid string, result *T) (*T, error) {
	collection, _, err := readContext(c)
	if err != nil {
		return nil, err
	}
	if err := GetReadCollection(collection).
		FindOne(c.Request.Context(),
			NewFilterBuilder().
				WithNotDeleteForCustomer(c).
				WithGUID(guid).
				Get()).
		Decode(result); err != nil {
		utils.LogNTraceError("failed to get document by id", err, c)
		return nil, err
	}
	return result, nil
}

func CountDocs(c *gin.Context, f bson.D) (int64, error) {
	collection, _, err := readContext(c)
	if err != nil {
		return 0, err
	}
	filter := NewFilterBuilder().
		WithNotDeleteForCustomer(c).
		WithFilter(f).
		Get()
	return GetReadCollection(collection).CountDocuments(c.Request.Context(), filter)
}

/////////////////////////////////////////mongo utils/////////////////////////////////////////
//delete document by id
func DeleteDoc(c *gin.Context, collection string, id string) error {
	_, err := GetWriteCollection(collection).DeleteOne(c.Request.Context(), bson.M{utils.ID_FIELD: id})
	return err
}

func Map2BsonD(m map[string]interface{}, keyPrefix string) bson.D {
	var result bson.D
	for k, v := range m {
		result = append(result, bson.E{Key: fmt.Sprintf("%s.%s", keyPrefix, k), Value: v})
	}
	return result
}

//helpers
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
	customerGUID = c.GetString(utils.CUSTOMER_GUID)
	if customerGUID == "" {
		err = fmt.Errorf("customerGUID is not in context")
	}
	return customerGUID, err
}

func readCollection(c *gin.Context) (collection string, err error) {
	collection = c.GetString(utils.COLLECTION)
	if collection == "" {
		err = fmt.Errorf("collection is not in context")
	}
	return collection, err
}
