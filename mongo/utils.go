package mongo

import (
	"fmt"
	"kubescape-config-service/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/////////////////////////////////////////gin handlers/////////////////////////////////////////

//HandleDeleteDoc gin handler for delete document by id in collection in context
func HandleDeleteDoc(c *gin.Context) {
	if !DocExists(c, c.Param(utils.GUID_FIELD)) {
		err := fmt.Errorf("document with id %s does not exist", c.Param(utils.GUID_FIELD))
		utils.LogNTraceError("handle delete document failed", err, c)
		c.AbortWithError(500, err)
		return
	}
	if err := DeleteDoc(c, c.GetString(utils.COLLECTION), c.Param(utils.GUID_FIELD)); err != nil {
		msg := fmt.Sprintf("failed to delete document GUID: %s  Collection: %s", c.Param(utils.GUID_FIELD), c.GetString(utils.COLLECTION))
		utils.LogNTraceError(msg, err, c)
		c.AbortWithError(500, err)
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

//////////////////////////////////Sugar functions for mongo using values in gic context ///////////////////////////////////////////

//GetAllForCustomer returns all not delete docs for customer from customerGUID and collection in context
func GetAllForCustomer[T any](c *gin.Context, result []T) ([]T, error) {
	return GetAllForCustomerWithProjection(c, result, nil)
}

//GetAllForCustomerWithProjection returns all not delete docs for customer from customerGUID and collection in context
func GetAllForCustomerWithProjection[T any](c *gin.Context, result []T, projection bson.D) ([]T, error) {
	collection := c.GetString(utils.COLLECTION)
	if collection == "" {
		return nil, fmt.Errorf("collection name is not in context")
	}
	customerGUID := c.GetString(utils.CUSTOMER_GUID)
	if customerGUID == "" {
		return nil, fmt.Errorf("customerGUID is not in context")
	}
	filter := NewFilterBuilder().WithNotDeleteForCustomer(c).Build()
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

//DCOExists returns true if document with _id exists for customer in context from collection in context
func DocExists(c *gin.Context, id string) bool {
	filter := NewFilterBuilder().
		WithNotDeleteForCustomer(c).
		WithValue(utils.ID_FIELD, id).
		Build()
	n, _ := GetReadCollection(c.GetString(utils.COLLECTION)).CountDocuments(c.Request.Context(), filter, options.Count().SetLimit(1))
	return n > 0
}

//GetDocByGUID returns document by GUID for customer in context from collection in context
func GetDocByGUID[T any](c *gin.Context, guid string, result *T) (*T, error) {
	collection := c.GetString(utils.COLLECTION)
	if collection == "" {
		return nil, fmt.Errorf("collection name is not in context")
	}
	customerGUID := c.GetString(utils.CUSTOMER_GUID)
	if customerGUID == "" {
		return nil, fmt.Errorf("customerGUID is not in context")
	}
	if err := GetReadCollection(collection).FindOne(c.Request.Context(),
		NewFilterBuilder().
			WithNotDeleteForCustomer(c).
			WithGUID(guid).
			Build()).Decode(result); err != nil {
		utils.LogNTraceError("failed to get document by id", err, c)
		return nil, err
	}
	return result, nil
}

/////////////////////////////////////////mongo utils/////////////////////////////////////////
//delete document by id
func DeleteDoc(c *gin.Context, collection string, id string) error {
	_, err := GetWriteCollection(collection).DeleteOne(c.Request.Context(), bson.M{utils.ID_FIELD: id})
	return err
}
