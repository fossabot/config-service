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
	customerConfigRouter := handlers.AddRoutes(g, handlers.NewRouterOptionsBuilder[*types.CustomerConfig]().
		WithPath(consts.CustomerConfigPath).
		WithDBCollection(consts.CustomerConfigCollection).
		WithServeGet(false).                          // customer config needs custom get handler
		WithServeDelete(false).                       // customer config needs custom delete handler
		WithValidatePutGUID(false).                   // customer config needs custom put validator
		WithPutValidators(validatePutCustomerConfig). //customer config custom put validator
		Get()...)

	customerConfigRouter.GET("", getCustomerConfigHandler)
	customerConfigRouter.DELETE("", deleteCustomerConfig)

	// add lazy cache to default customer config
	db.AddCachedDocument[*types.CustomerConfig](consts.DefaultCustomerConfigKey,
		consts.CustomerConfigCollection,
		db.NewFilterBuilder().WithGlobalNotDelete().WithName(consts.GlobalConfigName).Get(),
		time.Minute*5)
}
