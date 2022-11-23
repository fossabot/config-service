package types

import (
	"kubescape-config-service/utils/consts"
	"time"

	"github.com/armosec/armoapi-go/armotypes"
)

// Doc Content interface for data types embedded in DB documents
type DocContent interface {
	*Cluster | *PostureExceptionPolicy | *VulnerabilityExceptionPolicy
	InitNew()
	GetGUID() string
	SetGUID(guid string)
	GetName() string
	GetReadOnlyFields() []string
}

// redefine types for Doc Content implementations

type PostureExceptionPolicy armotypes.PostureExceptionPolicy
type Cluster armotypes.PortalCluster
type VulnerabilityExceptionPolicy armotypes.VulnerabilityExceptionPolicy

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
	return clusterReadOnlyFields
}
func (c *Cluster) InitNew() {
	c.SubscriptionDate = time.Now().UTC().Format(time.RFC3339)
	if c.Attributes == nil {
		c.Attributes = make(map[string]interface{})
	}
}

func (c *VulnerabilityExceptionPolicy) GetGUID() string {
	return c.GUID
}
func (c *VulnerabilityExceptionPolicy) SetGUID(guid string) {
	c.GUID = guid
}
func (c *VulnerabilityExceptionPolicy) GetName() string {
	return c.Name
}
func (c *VulnerabilityExceptionPolicy) GetReadOnlyFields() []string {
	return exceptionPolicyReadOnlyFields
}
func (c *VulnerabilityExceptionPolicy) InitNew() {
	c.CreationTime = time.Now().UTC().Format(time.RFC3339)
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
	return exceptionPolicyReadOnlyFields
}
func (p *PostureExceptionPolicy) InitNew() {
	p.CreationTime = time.Now().UTC().Format("time.RFC3339")
}

var commonReadOnlyFields = []string{consts.ID_FIELD, consts.NAME_FIELD, consts.GUID_FIELD}
var clusterReadOnlyFields = append([]string{"subscription_date"}, commonReadOnlyFields...)
var exceptionPolicyReadOnlyFields = append([]string{"creationTime"}, commonReadOnlyFields...)
var repositoryReadOnlyFields = append([]string{"creationDate", "provider", "owner", "repoName", "branchName"}, commonReadOnlyFields...)
