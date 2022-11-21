package mongo

import (
	"kubescape-config-service/types"

	uuid "github.com/satori/go.uuid"
)

type DocData interface {
	*types.Cluster | *types.PostureExceptionPolicy
	InitNew()
	GetGUID() string
	SetGUID(guid string)
	GetName() string
	GetReadOnlyFields() []string
}

type Document[T DocData] struct {
	ID        string   `json:"_id" bson:"_id"`
	Customers []string `json:"customers" bson:"customers,omitempty"`
	Content   T        `json:",inline" bson:"inline"`
}

/*func NewDocument[T DocData](data T, customerGUID string) (Document[T], error) {
	if cluster, ok := data.(*types.Cluster); ok {
		return NewClusterDocument(*cluster, customerGUID), nil
	}

}

func NewPostureExceptionDocument(e types.PostureExceptionPolicy, customerGUID string) Document[*types.PostureExceptionPolicy] {
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

func initNewCluster(c types.Cluster, customerGUID string) types.Cluster {
	if c.GUID == "" {
		c.GUID = uuid.NewV4().String()
	}
	if c.Attributes == nil {
		c.Attributes = make(map[string]interface{})
	}
	return c
}
*/
func NewDocument[T DocData](content T, customerGUID string) Document[T] {
	content.InitNew()
	if content.GetGUID() == "" {
		content.SetGUID(uuid.NewV4().String())
	}
	doc := Document[T]{
		ID:      content.GetGUID(),
		Content: content,
	}
	if customerGUID != "" {
		doc.Customers = append(doc.Customers, customerGUID)
	}
	return doc
}
