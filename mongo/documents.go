package mongo

import (
	"kubescape-config-service/types"
	"time"

	"github.com/armosec/armoapi-go/armotypes"
	uuid "github.com/satori/go.uuid"
)

type DocType interface {
	*types.Cluster | *armotypes.PostureExceptionPolicy
}

type Document[T DocType] struct {
	ID        string   `json:"_id" bson:"_id"`
	Customers []string `json:"customers" bson:"customers,omitempty"`
	Content   T        `json:",inline" bson:"inline"`
}

func NewPostureExceptionDocument(e armotypes.PostureExceptionPolicy, customerGUID string) Document[*armotypes.PostureExceptionPolicy] {
	if e.GUID == "" {
		e.GUID = uuid.NewV4().String()
	}
	if e.CreationTime == "" {
		e.CreationTime = time.Now().UTC().Format("2006-01-02T15:04:05.999")
	}
	return newDocument(e.GUID, customerGUID, &e)
}

func NewClusterDocument(c types.Cluster, customerGUID string) Document[*types.Cluster] {
	if c.GUID == "" {
		c.GUID = uuid.NewV4().String()
	}
	if c.Attributes == nil {
		c.Attributes = make(map[string]interface{})
	}
	return newDocument(c.GUID, customerGUID, &c)
}

func newDocument[T DocType](id, customerGUID string, content T) Document[T] {
	doc := Document[T]{
		ID:      id,
		Content: content,
	}
	if customerGUID != "" {
		doc.Customers = append(doc.Customers, customerGUID)
	}
	return doc
}
