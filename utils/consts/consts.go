package consts

const (
	COLLECTION                 = "collection"
	CUSTOMER_GUID              = "customerGUID"
	CUSTOMERS                  = "customers"
	CLUSTERS                   = "clusters"
	POSTURE_EXCEPTION_POLICIES = "v1_posture_exception_policies"

	ID_FIELD             = "_id"
	GUID_FIELD           = "portalbase.guid"
	NAME_FIELD           = "portalbase.name"
	DELETED_FIELD        = "is_deleted"
	ATTRIBUTES_FIELD     = "portalbase.attributes"
	SHORT_NAME_ATTRIBUTE = "alias"
	SHORT_NAME_FIELD     = ATTRIBUTES_FIELD + "." + SHORT_NAME_ATTRIBUTE
)
