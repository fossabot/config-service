package types

import (
	"config-service/utils/consts"
	"time"

	"github.com/armosec/armoapi-go/armotypes"
	uuid "github.com/satori/go.uuid"
)

// Document - document in db
type Document[T DocContent] struct {
	ID        string   `json:"_id" bson:"_id"`
	Customers []string `json:"customers" bson:"customers"`
	Content   T        `json:",inline" bson:"inline"`
}

// NewDocument - create new document per doc content T
func NewDocument[T DocContent](content T, customerGUID string) Document[T] {
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

// Doc Content interface for data types embedded in DB documents
type DocContent interface {
	*CustomerConfig | *Cluster | *PostureExceptionPolicy | *VulnerabilityExceptionPolicy | *Customer
	InitNew()
	GetName() string
	SetName(name string)
	GetReadOnlyFields() []string
	GetGUID() string
	SetGUID(guid string)
}

// redefine types for Doc Content implementations

// DocContent implementations
type CustomerConfig struct {
	armotypes.CustomerConfig `json:",inline" bson:"inline"`
	GUID                     string `json:"guid" bson:"guid"`
	CreationTime             string `json:"creationTime" bson:"creationTime"`
}

func (c *CustomerConfig) GetGUID() string {
	return c.GUID
}

func (c *CustomerConfig) SetGUID(guid string) {
	c.GUID = guid
}
func (c *CustomerConfig) GetName() string {
	if c.Name == "" &&
		c.Scope.Attributes != nil &&
		c.Scope.Attributes["cluster"] != "" {
		return c.Scope.Attributes["cluster"]
	}
	return c.Name
}
func (c *CustomerConfig) SetName(name string) {
	c.Name = name
}
func (c *CustomerConfig) GetReadOnlyFields() []string {
	return customerConfigReadOnlyFields
}
func (c *CustomerConfig) InitNew() {
	c.CreationTime = time.Now().UTC().Format(time.RFC3339)
	if c.Scope.Attributes != nil && c.Scope.Attributes["cluster"] != "" {
		c.Name = c.Scope.Attributes["cluster"]
	}
}

// DocContent implementations
type PostureExceptionPolicy armotypes.PostureExceptionPolicy
type Cluster armotypes.PortalCluster
type VulnerabilityExceptionPolicy armotypes.VulnerabilityExceptionPolicy
type Customer armotypes.PortalCustomer

func (c *Customer) GetGUID() string {
	return c.GUID
}
func (c *Customer) SetGUID(guid string) {
	c.GUID = guid
}
func (c *Customer) GetName() string {
	return c.Name
}
func (c *Customer) SetName(name string) {
	c.Name = name
}
func (c *Customer) GetReadOnlyFields() []string {
	return commonReadOnlyFields
}
func (c *Customer) InitNew() {
	c.SubscriptionDate = time.Now().UTC().Format(time.RFC3339)
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
func (c *Cluster) SetName(name string) {
	c.Name = name
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
func (c *VulnerabilityExceptionPolicy) SetName(name string) {
	c.Name = name
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
func (p *PostureExceptionPolicy) SetName(name string) {
	p.Name = name
}
func (p *PostureExceptionPolicy) GetReadOnlyFields() []string {
	return exceptionPolicyReadOnlyFields
}
func (p *PostureExceptionPolicy) InitNew() {
	p.CreationTime = time.Now().UTC().Format(time.RFC3339)
}

var commonReadOnlyFields = []string{consts.IdField, consts.NameField, consts.GUIDField}
var clusterReadOnlyFields = append([]string{"subscription_date"}, commonReadOnlyFields...)
var exceptionPolicyReadOnlyFields = append([]string{"creationTime"}, commonReadOnlyFields...)
var customerConfigReadOnlyFields = append([]string{"creationTime"}, commonReadOnlyFields...)

//var repositoryReadOnlyFields = append([]string{"creationDate", "provider", "owner", "repoName", "branchName"}, commonReadOnlyFields...)
