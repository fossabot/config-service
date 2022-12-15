package handlers

type queryParamsConfig struct {
	params2Query   map[string]queryConfig
	defaultContext string
}

type queryConfig struct {
	fieldName   string
	pathInArray string
	isArray     bool
}

func defaultConfig() *queryParamsConfig {
	return &queryParamsConfig{
		defaultContext: "attributes",
		params2Query: map[string]queryConfig{
			"attributes": {
				fieldName:   "attributes",
				pathInArray: "",
				isArray:     false,
			},
		},
	}
}

func GetPostureExceptionQueryConfig() *queryParamsConfig {
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

func GetVulnerabilityExceptionConfig() *queryParamsConfig {
	config := defaultConfig()
	config.defaultContext = "designators"
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
		fieldName:   "vulnerabilities",
		pathInArray: "",
		isArray:     true,
	}
	return config
}
