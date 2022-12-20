# config-service


1. [Overview](#overview)
2. [Document types](#document-types)
3. [Generic Handlers](#generic-handlers)
4. [Adding new document type handler](#adding-new-document-type-handler)
    1. [Todo List](#todo-list)
    2. [Using the generic handlers](#using-the-generic-handlers)
    3. [Router options](#router-options)
    4. [Customized behavior](#customized-behavior)




## Overview

The config service is a CRUD service of configuration data for kubescape.

It uses [gin web framework](https://github.com/gin-gonic/gin) for http handling and  [mongo](https://github.com/mongodb/mongo-go-driver) as the service data base.

The config service provides DB package for generic documents handling operations and handlers package for common handling.

![Alt text](docs/overview.jpg?raw=true "Overview")

## Document types
The service serve documents of DocContent type parameter, each document type should be added to the Doc content types constraint and implement the DocContent interface.
Both the db and handlers functions are parameterized by the DocContent type and do not need type casting.

```go
clusters := []*types.Cluster{}
frameworks := []*types.Framework{}
//method returns array of specified type 
clusters, err = db.GetAllForCustomer[*types.Cluster](...)
frameworks, err = db.GetAllForCustomer[*types.Framework](...)
```
For more details see the [types package](/types/types.go) section. 

![Alt text](docs/types.jpg?raw=true "Types")

## Generic Handlers
The generic handlers uses data in the gin context that were set by other middleware or handlers to perform the desired action like DocContent that was sent in the request body customer by the Post/Put handlers, customer guid set by the authenticate middleware, request logger set by the logger middleware, db collection name set by the db middleware and so on.
For full list ok context keys see [const.go](/utils/consts/const.go).


## Adding new document type handler
### Todo List
1. Add your new type to the DocContent types constraint in the [types package](/types/types.go) and implement the DocContent interface.
2. Add the strings of the new type path and DB collection in [const.go](/utils/consts/const.go) for the new type.
3. Add a folder under the routes folder for the new type and add a new file with ```func AddRoutes(g *gin.Engine) ``` function that adds handlers for the new type.
4. call the ```AddRoutes``` function in the [main.go](/main.go) file after the authentication middleware.

### Using the generic handlers
Endpoint handlers can configure the endpoint behavior and add handlers using ```handlers.AddRoutes ``` function with desired options 


```go
package "myType"
import (
	"config-service/handlers"
	"config-service/types"
	"config-service/utils/consts"

	"github.com/gin-gonic/gin"
)

function AddRoutes(g *gin.Engine) {
    //create a new router for myType endpoint with default options 
    routerOptions := handlers.NewRouterOptionsBuilder[*types.MyType].
    WithPath("/myType").
    WithDBCollection(consts.MyTypeCollection).
    //add get by name e.g. GET /myType?typeName="x"
    WithNameQuery(consts.MyTypeNameQueryParam).
    //remove get name list e.g. GET /myType?list
    WithGetNamesList(false)
    //add handlers for myType endpoint
    handlers.AddRoutes(g, routerOptions.Get()...)
}
```
#### Router options
| Method/Action | Description | Option setting example | Default |
| ------------- | ----------- | -------- | -------- | 
|GET all  | get all user's documents | routerOptions.WithServeGet(true) | On |
|GET list of names  | get list of documents names if "list" query param is set (e.g. GET /myType?list) |  routerOptions.WithGetNamesList(true) | On
|GET all with global  | get all user's and global (no owner) documents | routerOptions.WithIncludeGlobalDocs(true) | Off |
|GET by name  | get a document by name using query param (e.g. GET /myType?typeName="x") |  routerOptions.WithNameQuery("typeName") | Off
|GET by query  | get a document by query params according to given [query config](handlers/scopequery.go) (e.g. GET /myType?scope.cluster="nginx") |  routerOptions.WithQueryConfig(&queryConfig) | Off |
|POST with guid in path or body | create a new document, the post operation can be configured with additional customized or predefined [validators](handlers/validate.go) like unique name, unique short name attribute or custom validator   |  routerOptions.WithServePost(true).WithValidatePostUniqueName(true).WithPostValidator(myValidator) | On with unique name validator
|PUT  | update a document, the put operation can be configured with additional customized or predefined [validators](handlers/validate.go) like GUID existence , or custom validator   |  routerOptions.WithServePut(true).WithValidatePutGUID(true).WithPutValidator(myValidator) | On with guid existence validator
|DELETE with guid in path | delete a document   |  routerOptions.WithServeDelete(true) | On
|DELETE by name  | delete a document by name   |  routerOptions.WithDeleteByName(true) | Off

### Customized behavior
Endpoint that requirers customized behavior for some routes can still use the generic ```handlers.AddRoutes ``` for the rest of the routes and add custom handlers for the specific routes or add custom validators for the generic handlers, see [customer configuration](routes/customer_config/routes.go) for example.



## Testing
The service main test create a [testify suite](suite_test.go) that runs a mongo container and the config service for end to end testing.
Endpoint that are served using the generic handlers can reuse the [common tests functions](testers_test.go) to test the endpoint behavior.
For details see the existing [endpoint tests](service_test.go) 









