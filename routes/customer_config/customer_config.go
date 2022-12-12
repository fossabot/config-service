package customer_config

import (
	"config-service/dbhandler"
	"config-service/types"
	"config-service/utils/consts"
	"net/http"

	"github.com/imdario/mergo"

	"github.com/gin-gonic/gin"
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
		dbhandler.ResponseInternalServerError(c, "failed to get default config", err)
		return true
	}

	//case default config is requested - return it
	if configName == consts.GlobalConfigName {
		c.JSON(http.StatusOK, defaultConfig)
		return true
	}
	//try and get config by name from db
	doc, err := dbhandler.GetDocByName[types.CustomerConfig](c, configName)
	if err != nil {
		dbhandler.ResponseInternalServerError(c, "failed to get document by name", err)
		return true
	} else if unmerged, _ := c.GetQuery("unmerged"); unmerged != "" {
		//case unmerged is requested - return the unmerged config if exists
		if doc == nil {
			dbhandler.ResponseDocumentNotFound(c)
			return true
		} else {
			c.JSON(http.StatusOK, doc)
			return true
		}
	}
	//case customer config is requested - return it merged with default config
	if configName == consts.CustomerConfigName {
		if doc != nil {
			if err := mergo.Merge(doc, *defaultConfig); err != nil {
				dbhandler.ResponseInternalServerError(c, "failed to merge configuration", err)
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
	customerConfig, err := dbhandler.GetDocByName[types.CustomerConfig](c, consts.CustomerConfigName)
	if err != nil {
		dbhandler.ResponseInternalServerError(c, "failed to get document by name", err)
		return true
	}
	customerConfig, err = mergeConfigurations(customerConfig, defaultConfig)
	if err != nil {
		dbhandler.ResponseInternalServerError(c, "failed to merge configuration", err)
		return true
	}

	doc, err = mergeConfigurations(doc, customerConfig)
	if err != nil {
		dbhandler.ResponseInternalServerError(c, "failed to merge configuration", err)
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
		dbhandler.ResponseFailedToBindJson(c, err)
		return
	}
	configName := getConfigName(c)
	if configName == "" {
		if doc.Name != "" {
			configName = doc.Name
		} else {
			dbhandler.ResponseMissingName(c)
			return
		}
	}
	if existingDoc, err := dbhandler.GetDocByName[types.CustomerConfig](c, configName); err != nil {
		dbhandler.ResponseInternalServerError(c, "failed to get doc by name", err)
		return
	} else if existingDoc == nil {
		dbhandler.ResponseDocumentNotFound(c)
		return
	} else {
		doc.SetGUID(existingDoc.GetGUID())
	}
	c.Set(consts.DocContentKey, doc)
	c.Next()
}

func deleteCustomerConfig(c *gin.Context) {
	if configName := getConfigName(c); configName != "" {
		if configName == consts.GlobalConfigName {
			dbhandler.ResponseBadRequest(c, "default config cannot be deleted")
			return
		}
		dbhandler.DeleteDocByNameHandler[*types.CustomerConfig](c, configName)
	} else {
		dbhandler.ResponseMissingName(c)
		return
	}
}

func getConfigName(c *gin.Context) string {
	if clusterName, ok := c.GetQuery(consts.ClusterNameParam); ok && clusterName != "" {
		return clusterName
	}
	if configName, ok := c.GetQuery(consts.ConfigNameParam); ok && configName != "" {
		return configName
	}
	if scope, ok := c.GetQuery(consts.ScopeParam); ok && scope != "" {
		switch scope {
		case consts.CustomerScope:
			return consts.CustomerConfigName
		case consts.DefaultScope:
			return consts.GlobalConfigName
		}
	}
	return ""
}
