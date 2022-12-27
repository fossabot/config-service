package handlers

type QueryParamsConfig struct {
	Params2Query   map[string]QueryConfig
	DefaultContext string
}

type QueryConfig struct {
	FieldName   string
	PathInArray string
	IsArray     bool
}

func DefaultQueryConfig() *QueryParamsConfig {
	return &QueryParamsConfig{
		DefaultContext: "attributes",
		Params2Query: map[string]QueryConfig{
			"attributes": {
				FieldName:   "attributes",
				PathInArray: "",
				IsArray:     false,
			},
		},
	}
}
