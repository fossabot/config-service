package types

import "github.com/armosec/armoapi-go/armotypes"

type Base armotypes.PortalBase

//TODO move the external repo for reuse with other projects
/*type Base struct {
	GUID       string                 `json:"guid"  bson:"guid,omitempty"`
	Name       string                 `json:"name"  bson:"name,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty" bson:"attributes,omitempty"`
}*/

type Cluster struct {
	armotypes.PortalBase `json:",inline" bson:"inline"`
	SubscriptionDate     string `json:"subscription_date,omitempty" bson:"subscription_date,omitempty"`
	LastLoginDate        string `json:"last_login_date,omitempty" bson:"last_login_date,omitempty"`
}

//type PostureExceptionPolicy armotypes.PostureExceptionPolicy
