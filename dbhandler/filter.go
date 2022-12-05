package dbhandler

import (
	"kubescape-config-service/utils/consts"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// FilterBuilder builds filters for queries
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

func (f *FilterBuilder) WithGlobalNotDelete() *FilterBuilder {
	return f.WithValue(consts.CustomersField, "").WithNotDeleted()
}

func (f *FilterBuilder) WithNotDeleteForCustomerAndGlobal(c *gin.Context) *FilterBuilder {
	return f.WithCustomerAndGlobal(c).WithNotDeleted()
}

func (f *FilterBuilder) WithGUID(guid string) *FilterBuilder {
	return f.WithValue(consts.GUIDField, guid)
}

func (f *FilterBuilder) WithID(id string) *FilterBuilder {
	return f.WithValue(consts.IdField, id)
}

func (f *FilterBuilder) WithName(name string) *FilterBuilder {
	return f.WithValue(consts.NameField, name)
}

func (f *FilterBuilder) WithCustomer(c *gin.Context) *FilterBuilder {
	return f.WithValue(consts.CustomersField, c.GetString(consts.CustomerGUID))
}

func (f *FilterBuilder) WithCustomerAndGlobal(c *gin.Context) *FilterBuilder {
	return f.WithIn(consts.CustomersField, []string{c.GetString(consts.CustomerGUID), ""})
}

func (f *FilterBuilder) WithNotDeleted() *FilterBuilder {
	return f.WithNotEqual(consts.DeletedField, true)
}

func (f *FilterBuilder) WithDeleted() *FilterBuilder {
	return f.WithValue(consts.DeletedField, true)
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

func (f *FilterBuilder) WarpElementMatch() *FilterBuilder {
	f.filter = bson.D{{Key: "$elemMatch", Value: f.filter}}
	return f
}

func (f *FilterBuilder) WarpOr() *FilterBuilder {
	m := bson.M{}
	for i := range f.filter {
		m[f.filter[i].Key] = f.filter[i].Value

	}
	f.filter = bson.D{{Key: "$or", Value: bson.A{m}}}
	return f
}

func (f *FilterBuilder) WarpWithField(field string) *FilterBuilder {
	f.filter = bson.D{{Key: field, Value: f.filter}}
	return f
}
