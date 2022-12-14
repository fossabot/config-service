package customer_config

import (
	"config-service/db"
	"config-service/handlers"
	"config-service/types"
	"config-service/utils/consts"
	"time"

	"github.com/gin-gonic/gin"
)

func AddRoutes(g *gin.Engine) {
	customerConfig := g.Group(consts.CustomerConfigPath)
	customerConfig.Use(handlers.DBContextMiddleware(consts.CustomerConfigCollection))

	customerConfig.GET("", getCustomerConfigHandler)
	customerConfig.POST("", handlers.HandlePostDocWithUniqueNameValidation[*types.CustomerConfig]()...)
	customerConfig.PUT("", handlers.HandlePutDocWithValidation(validatePutCustomerConfig)...)
	customerConfig.PUT("/:"+consts.GUIDField, handlers.HandlePutDocWithGUIDValidation[*types.CustomerConfig]()...)
	customerConfig.DELETE("", deleteCustomerConfig)

	db.AddCachedDocument[*types.CustomerConfig](consts.DefaultCustomerConfigKey,
		consts.CustomerConfigCollection,
		db.NewFilterBuilder().WithGlobalNotDelete().WithName(consts.GlobalConfigName).Get(),
		time.Minute*5)
}
