package customer_config

import (
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"
	"kubescape-config-service/utils/log"
	"net/http"

	"github.com/imdario/mergo"

	"github.com/gin-gonic/gin"
)

const (
	globalConfigName   = "default"
	customerConfigName = "CustomerConfig"
	clusterNameParam   = "clusterName"
	configNameParam    = "configName"
	scopeParam         = "scope"
	customerScope      = "customer"
	defaultScope       = "default"
)

func getCustomerConfigHandler(c *gin.Context) {
	if dbhandler.GetNamesListHandler[*types.CustomerConfig](c, true) {
		return
	}
	if getCustomerConfigByNameHandler(c) {
		return
	}
	dbhandler.HandleGetAllWithGlobals[*types.CustomerConfig](c)
}

func getCustomerConfigByNameHandler(c *gin.Context) bool {
	configName := getConfigName(c)
	if configName == "" {
		return false
	}
	//get the default config
	defaultConfig, err := dbhandler.GetCachedDocument[*types.CustomerConfig](consts.DefaultCustomerConfigKey)
	if err != nil {
		log.LogNTraceError("failed to get default config", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return true
	}

	//case default config is requested - return it
	if configName == globalConfigName {
		c.JSON(http.StatusOK, defaultConfig)
		return true
	}
	//try and get config by name from db
	doc, err := dbhandler.GetDocByName[types.CustomerConfig](c, configName)
	if err != nil {
		log.LogNTraceError("failed to get doc", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return true
	} else if unmerged, _ := c.GetQuery("unmerged"); unmerged != "" {
		//case unmerged is requested - return the unmerged config if exists
		if doc == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return true
		} else {
			c.JSON(http.StatusOK, doc)
			return true
		}
	}
	//case customer config is requested - return it merged with default config
	if configName == customerConfigName {
		if doc != nil {
			if err := mergo.Merge(doc, *defaultConfig); err != nil {
				log.LogNTraceError("failed to merge config", err, c)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return true
			} else {
				c.JSON(http.StatusOK, doc)
				return true
			}
		} else {
			//case customer config is requested but not exists - return default config
			c.JSON(http.StatusOK, defaultConfig)
			return true
		}
	}
	//case cluster config is requested - return it merged with customer and default config
	customerConfig, err := dbhandler.GetDocByName[types.CustomerConfig](c, customerConfigName)
	if err != nil {
		log.LogNTraceError("failed to get doc", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return true
	}
	customerConfig, err = mergeConfigurations(customerConfig, defaultConfig)
	if err != nil {
		log.LogNTraceError("failed to merge config", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return true
	}

	doc, err = mergeConfigurations(doc, customerConfig)
	if err != nil {
		log.LogNTraceError("failed to merge config", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return true
	}
	c.JSON(http.StatusOK, doc)
	return true
}

func mergeConfigurations(dest, src *types.CustomerConfig) (*types.CustomerConfig, error) {
	if dest == nil {
		return src, nil
	}
	if src == nil {
		return dest, nil
	}
	destCopy := *dest
	if err := mergo.Merge(&destCopy, *src); err != nil {
		return nil, err
	}
	return &destCopy, nil
}

func putCustomerConfigValidation(c *gin.Context) {
	var doc *types.CustomerConfig
	if err := c.ShouldBindJSON(&doc); err != nil {
		log.LogNTraceError("failed to bind json", err, c)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	configName := getConfigName(c)
	if configName == "" {
		if doc.Name != "" {
			configName = doc.Name
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "configName is required"})
			return
		}
	}
	if existingDoc, err := dbhandler.GetDocByName[types.CustomerConfig](c, configName); err != nil {
		log.LogNTraceError("failed to get doc", err, c)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if existingDoc == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"error": "document not found"}})
		return
	} else {
		doc.SetGUID(existingDoc.GetGUID())
	}
	c.Set(consts.DocContentKey, doc)
	c.Next()
}

func deleteCustomerConfig(c *gin.Context) {
	if configName := getConfigName(c); configName != "" {
		if configName == globalConfigName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "default config cannot be deleted"})
			return
		}
		dbhandler.DeleteDocByNameHandler[*types.CustomerConfig](c, configName)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "configName is required"})
	}
}

func getConfigName(c *gin.Context) string {
	if clusterName, ok := c.GetQuery(clusterNameParam); ok && clusterName != "" {
		return clusterName
	}
	if configName, ok := c.GetQuery(configNameParam); ok && configName != "" {
		return configName
	}
	if scope, ok := c.GetQuery(scopeParam); ok && scope != "" {
		switch scope {
		case customerScope:
			return customerConfigName
		case defaultScope:
			return globalConfigName
		}
	}
	return ""
}
