package types

import (
	"kubescape-config-service/utils"
	"time"

	"github.com/armosec/armoapi-go/armotypes"
)

var CommonROFields = []string{utils.ID_FIELD, utils.GUID_FIELD, utils.NAME_FIELD}
var ClusterROFields = append([]string{"subscription_date"}, CommonROFields...)
var PostureExceptionROFields = append([]string{"creationTime"}, CommonROFields...)
var RepositoryROFields = append([]string{"creationDate", "provider", "owner", "repoName", "branchName"}, CommonROFields...)

// "creationTime", "creationDate", "provider", "owner", "repoName", "branchName"}

type Cluster struct {
	armotypes.PortalBase `json:",inline" bson:"inline"`
	SubscriptionDate     string `json:"subscription_date,omitempty" bson:"subscription_date,omitempty"`
	LastLoginDate        string `json:"last_login_date,omitempty" bson:"last_login_date,omitempty"`
}

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
		c.SubscriptionDate = time.Now().UTC().Format("2006-01-02T15:04:05.999")
	}
	if c.Attributes == nil {
		c.Attributes = make(map[string]interface{})
	}
}

type PostureExceptionPolicy armotypes.PostureExceptionPolicy

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
		p.CreationTime = time.Now().UTC().Format("2006-01-02T15:04:05.999")
	}
}
