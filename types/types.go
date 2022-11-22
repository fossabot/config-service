package types

import (
	"kubescape-config-service/utils/consts"
	"time"

	"github.com/armosec/armoapi-go/armotypes"
)

// Doc Content interface for data types embedded in DB documents
type DocContent interface {
	*Cluster | *PostureExceptionPolicy
	InitNew()
	GetGUID() string
	SetGUID(guid string)
	GetName() string
	GetReadOnlyFields() []string
}

// redefine types for Doc Content implementations

type PostureExceptionPolicy armotypes.PostureExceptionPolicy

// TODO move to armotypes
type Cluster armotypes.PortalCluster

func (c *Cluster) GetGUID() string {
	return c.GUID
}
func (c *Cluster) SetGUID(guid string) {
	c.GUID = guid
}
func (c *Cluster) GetName() string {
	return c.Name
}
func (c *Cluster) GetReadOnlyFields() []string {
	return ClusterROFields
}
func (c *Cluster) InitNew() {
	if c.SubscriptionDate == "" {
		c.SubscriptionDate = time.Now().UTC().Format(time.RFC3339)
	}
	if c.Attributes == nil {
		c.Attributes = make(map[string]interface{})
	}
}

func (p *PostureExceptionPolicy) GetGUID() string {
	return p.GUID
}
func (p *PostureExceptionPolicy) SetGUID(guid string) {
	p.GUID = guid
}
func (p *PostureExceptionPolicy) GetName() string {
	return p.Name
}
func (p *PostureExceptionPolicy) GetReadOnlyFields() []string {
	return PostureExceptionROFields
}
func (p *PostureExceptionPolicy) InitNew() {
	if p.CreationTime == "" {
		p.CreationTime = time.Now().UTC().Format("time.RFC3339")
	}
}

var CommonROFields = []string{consts.ID_FIELD, consts.NAME_FIELD, consts.GUID_FIELD}
var ClusterROFields = append([]string{"subscription_date"}, CommonROFields...)
var PostureExceptionROFields = append([]string{"creationTime"}, CommonROFields...)
var RepositoryROFields = append([]string{"creationDate", "provider", "owner", "repoName", "branchName"}, CommonROFields...)
