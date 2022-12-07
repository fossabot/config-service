package consts

const (
	CustomerGUID = "customerGUID"
	Collection   = "collection"

	//PATHS
	ClusterPath                      = "/cluster"
	PostureExceptionPolicyPath       = "/v1_posture_exception_policy"
	VulnerabilityExceptionPolicyPath = "/v1_vulnerability_exception_policy"
	CustomerConfigPath               = "/v1_customer_configuration"

	//DB collections
	ClustersCollection                     = "clusters"
	PostureExceptionPolicyCollection       = "v1_posture_exception_policies"
	VulnerabilityExceptionPolicyCollection = "v1_vulnerability_exception_policies"
	CustomerConfigCollection               = "v1_customer_configurations"

	//Common document fields
	IdField           = "_id"
	GUIDField         = "guid"
	NameField         = "name"
	DeletedField      = "is_deleted"
	AttributesField   = "attributes"
	ShotNameAttribute = "alias"
	ShotNameField     = AttributesField + "." + ShotNameAttribute
	CustomersField    = "customers"

	//Query params
	ListParam       = "list"
	PolicyNameParam = "policyName"

	//Context keys
	DocContentKey = "docContent"

	//Cached documents keys
	DefaultCustomerConfigKey = "defaultCustomerConfig"

	//customer configuration fields
	GlobalConfigName   = "default"
	CustomerConfigName = "CustomerConfig"
	ClusterNameParam   = "clusterName"
	ConfigNameParam    = "configName"
	ScopeParam         = "scope"
	CustomerScope      = "customer"
	DefaultScope       = "default"
)
