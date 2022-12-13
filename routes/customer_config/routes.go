package customer_config

import (
	"config-service/dbhandler"
	"config-service/types"
	"config-service/utils/consts"
	"time"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	customerConfig := g.Group(consts.CustomerConfigPath)
	customerConfig.Use(dbhandler.DBContextMiddleware(consts.CustomerConfigCollection))

	customerConfig.GET("", getCustomerConfigHandler)
	customerConfig.POST("", dbhandler.HandlePostDocWithUniqueNameValidation[*types.CustomerConfig]()...)
	customerConfig.PUT("", dbhandler.HandlePutDocWithValidation(validatePutCustomerConfig)...)
	customerConfig.PUT("/:"+consts.GUIDField, dbhandler.HandlePutDocWithGUIDValidation[*types.CustomerConfig]()...)
	customerConfig.DELETE("", deleteCustomerConfig)

	dbhandler.AddCachedDocument[*types.CustomerConfig](consts.DefaultCustomerConfigKey,
		consts.CustomerConfigCollection,
		dbhandler.NewFilterBuilder().WithGlobalNotDelete().WithName(consts.GlobalConfigName).Get(),
		time.Minute*5)
}
