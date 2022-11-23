package dbhandler

type scopeParamsConfig struct {
	params2Query map[string]queryConfig
}

type queryConfig struct {
	fieldName   string
	pathInArray string
	isArray     bool
}

func defaultConfig() *scopeParamsConfig {
	return &scopeParamsConfig{
		params2Query: map[string]queryConfig{
			"attributes": {
				fieldName:   "attributes",
				pathInArray: "",
				isArray:     false,
			},
		},
	}
}

func GetPostureExceptionQueryConfig() *scopeParamsConfig {
	config := defaultConfig()
	config.params2Query["scope"] = queryConfig{
		fieldName:   "resources",
		pathInArray: "attributes",
		isArray:     true,
	}
	config.params2Query["resources"] = queryConfig{
		fieldName:   "resources",
		pathInArray: "",
		isArray:     true,
	}
	config.params2Query["posturePolicies"] = queryConfig{
		fieldName:   "posturePolicies",
		pathInArray: "",
		isArray:     true,
	}
	return config
}

func GetVulnerabilityExceptionConfig() *scopeParamsConfig {
	config := defaultConfig()
	config.params2Query["scope"] = queryConfig{
		fieldName:   "designators",
		pathInArray: "attributes",
		isArray:     true,
	}
	config.params2Query["designators"] = queryConfig{
		fieldName:   "designators",
		pathInArray: "",
		isArray:     true,
	}
	config.params2Query["vulnerabilities"] = queryConfig{
		fieldName:   "designators",
		pathInArray: "attributes",
		isArray:     true,
	}
	return config
}
