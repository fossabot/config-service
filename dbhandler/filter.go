package dbhandler

import (
	"kubescape-config-service/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type FilterBuilder struct {
	filter bson.D
}

func NewFilterBuilder() *FilterBuilder {
	return &FilterBuilder{
		filter: bson.D{},
	}
}

func (f *FilterBuilder) WithFilter(filter bson.D) *FilterBuilder {
	for e := range filter {
		f.filter = append(f.filter, filter[e])
	}
	return f
}

func (f *FilterBuilder) Get() bson.D {
	return f.filter
}

func (f *FilterBuilder) WithNotDeleteForCustomer(c *gin.Context) *FilterBuilder {
	return f.WithCustomer(c).WithNotDeleted()
}

func (f *FilterBuilder) WithGUID(guid string) *FilterBuilder {
	return f.WithValue(utils.GUID_FIELD, guid)
}

func (f *FilterBuilder) WithID(id string) *FilterBuilder {
	return f.WithValue(utils.ID_FIELD, id)
}

func (f *FilterBuilder) WithName(name string) *FilterBuilder {
	return f.WithValue(utils.NAME_FIELD, name)
}

func (f *FilterBuilder) WithCustomer(c *gin.Context) *FilterBuilder {
	return f.WithValue(utils.CUSTOMERS, c.GetString(utils.CUSTOMER_GUID))
}

func (f *FilterBuilder) WithNotDeleted() *FilterBuilder {
	return f.WithNotEqual(utils.DELETED_FIELD, true)
}

func (f *FilterBuilder) WithDeleted() *FilterBuilder {
	return f.WithValue(utils.DELETED_FIELD, true)
}

func (f *FilterBuilder) WithValue(key string, value interface{}) *FilterBuilder {
	f.filter = append(f.filter, bson.E{Key: key, Value: value})
	return f
}

func (f *FilterBuilder) WithNotEqual(key string, value interface{}) *FilterBuilder {
	f.filter = append(f.filter, bson.E{Key: key, Value: bson.D{{Key: "$ne", Value: value}}})
	return f
}

func (f *FilterBuilder) WithIn(key string, value interface{}) *FilterBuilder {
	f.filter = append(f.filter, bson.E{Key: key, Value: bson.D{{Key: "$in", Value: value}}})
	return f
}

func (f *FilterBuilder) WithNotIn(key string, value interface{}) *FilterBuilder {
	f.filter = append(f.filter, bson.E{Key: key, Value: bson.D{{Key: "$nin", Value: value}}})
	return f
}

func (f *FilterBuilder) WithExists(key string, value bool) *FilterBuilder {
	f.filter = append(f.filter, bson.E{Key: key, Value: bson.D{{Key: "$exists", Value: value}}})
	return f
}

func (f *FilterBuilder) AddNotExists(key string) *FilterBuilder {
	f.filter = append(f.filter, bson.E{Key: key, Value: bson.D{{Key: "$exists", Value: false}}})
	return f
}
