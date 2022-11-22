package dbhandler

import (
	"kubescape-config-service/types"

	uuid "github.com/satori/go.uuid"
)

// Document - db document that contains DocContent data
type Document[T types.DocContent] struct {
	ID        string   `json:"_id" bson:"_id"`
	Customers []string `json:"customers" bson:"customers,omitempty"`
	Content   T        `json:",inline" bson:"inline"`
}

// NewDocument - create new document per doc content T
func NewDocument[T types.DocContent](content T, customerGUID string) Document[T] {
	content.InitNew()
	content.SetGUID(uuid.NewV4().String())

	doc := Document[T]{
		ID:      content.GetGUID(),
		Content: content,
	}
	if customerGUID != "" {
		doc.Customers = append(doc.Customers, customerGUID)
	}
	return doc
}
