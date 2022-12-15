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
	customerConfigRouter := handlers.AddRoutes[*types.CustomerConfig](
		g,
		handlers.WithPath[*types.CustomerConfig](consts.CustomerConfigPath),
		handlers.WithDBCollection[*types.CustomerConfig](consts.CustomerConfigCollection),
		handlers.WithServeGet[*types.CustomerConfig](false),        // customer config needs custom get handler
		handlers.WithServeDelete[*types.CustomerConfig](false),     // customer config needs custom delete handler
		handlers.WithValidatePutGUID[*types.CustomerConfig](false), // customer config needs custom put validator
		handlers.WithPutValidators(validatePutCustomerConfig),      //customer config custom put validator

	)
	customerConfigRouter.GET("", getCustomerConfigHandler)
	customerConfigRouter.DELETE("", deleteCustomerConfig)

	// add lazy cache to default customer config
	db.AddCachedDocument[*types.CustomerConfig](consts.DefaultCustomerConfigKey,
		consts.CustomerConfigCollection,
		db.NewFilterBuilder().WithGlobalNotDelete().WithName(consts.GlobalConfigName).Get(),
		time.Minute*5)
}
