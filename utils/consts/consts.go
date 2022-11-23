package consts

const (
	CUSTOMER_GUID = "customerGUID"
	COLLECTION    = "collection"

	//PATHS
	CLUSTER_PATH                        = "/cluster"
	POSTURE_EXCEPTION_POLICY_PATH       = "/v1_posture_exception_policy"
	VULNERABILITY_EXCEPTION_POLICY_PATH = "/v1_vulnerability_exception_policy"

	//DB collections
	CLUSTERS_COLLECTION                         = "clusters"
	POSTURE_EXCEPTION_POLICIES_COLLECTION       = "v1_posture_exception_policies"
	VULNERABILITY_EXCEPTION_POLICIES_COLLECTION = "v1_vulnerability_exception_policies"

	//Common document fields
	ID_FIELD             = "_id"
	GUID_FIELD           = "guid"
	NAME_FIELD           = "name"
	DELETED_FIELD        = "is_deleted"
	ATTRIBUTES_FIELD     = "attributes"
	SHORT_NAME_ATTRIBUTE = "alias"
	SHORT_NAME_FIELD     = ATTRIBUTES_FIELD + "." + SHORT_NAME_ATTRIBUTE
	CUSTOMERS_FIELD      = "customers"

	//Query params
	LIST_PARAM        = "list"
	POLICY_NAME_PARAM = "policyName"

	//Context keys
	DOC_CONTENT_KEY = "docContent"
)
