package customer_config

import (
	"kubescape-config-service/dbhandler"
	"kubescape-config-service/types"
	"kubescape-config-service/utils/consts"
	"time"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	customerConfig := g.Group(consts.CustomerConfigPath)
	customerConfig.Use(dbhandler.DBContextMiddleware(consts.CustomerConfigCollection))

	customerConfig.GET("/", getCustomerConfigHandler)
	customerConfig.POST("/", dbhandler.HandlePostDocWithValidation[*types.CustomerConfig]()...)
	customerConfig.PUT("/", putCustomerConfigValidation, dbhandler.HandlePutDocFromContext[*types.CustomerConfig])
	customerConfig.PUT("/:"+consts.GUIDField, dbhandler.HandlePutDocWithValidation[*types.CustomerConfig]()...)
	customerConfig.DELETE("/", deleteCustomerConfig)

	dbhandler.AddCachedDocument[*types.CustomerConfig](consts.DefaultCustomerConfigKey,
		consts.CustomerConfigCollection,
		dbhandler.NewFilterBuilder().WithGlobalNotDelete().WithName(globalConfigName).Get(),
		time.Minute*5)
}
