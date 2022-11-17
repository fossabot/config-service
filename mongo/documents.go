package mongo

import (
	"kubescape-config-service/types"
	"time"

	uuid "github.com/satori/go.uuid"
)

type Document interface {
	GetCollectionName() string
}

type BaseDocument struct {
	ID string `json:"_id" bson:"_id"`
}

func NewClusterDocument(c types.Cluster) ClusterDoc {
	if c.GUID == "" {
		c.GUID = uuid.NewV4().String()
	}
	if c.SubscriptionDate == "" {
		time.Now().UTC().Format("2006-01-02T15:04:05.999")
		c.SubscriptionDate = time.Now().UTC().Format("2006-01-02T15:04:05.999")
	}
	if c.Attributes == nil {
		c.Attributes = make(map[string]interface{})
	}
	return ClusterDoc{
		BaseDocument: BaseDocument{
			ID: c.GUID,
		},
		Cluster: c,
	}
}

type ClusterDoc struct {
	BaseDocument  `json:",inline" bson:"inline"`
	types.Cluster `json:",inline" bson:"inline"`
	Customers     []string `json:"customers" bson:"customers,omitempty"`
}

func (c *ClusterDoc) ClearReadOnlyFields() {
	c.Cluster.SubscriptionDate = ""
	c.Cluster.LastLoginDate = ""
	c.Cluster.Name = ""
	c.Cluster.GUID = ""
	c.Customers = nil
}
