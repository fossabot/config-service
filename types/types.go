package types

type Base struct {
	GUID       string                 `json:"guid"  bson:"guid,omitempty"`
	Name       string                 `json:"name"  bson:"name,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty" bson:"attributes,omitempty"`
}

type Cluster struct {
	Base             `json:",inline" bson:"inline"`
	SubscriptionDate string `json:"subscription_date,omitempty" bson:"subscription_date,omitempty"`
	LastLoginDate    string `json:"last_login_date,omitempty" bson:"last_login_date,omitempty"`
}

//Old Types names for backward compatibility
type PortalBase Base
type PortalCluster Cluster

func (d *Base) MergeAttributes(source *Base) {
	for k, v := range source.Attributes {
		if _, ok := d.Attributes[k]; !ok {
			d.Attributes[k] = v
		}
	}
}
