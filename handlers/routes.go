package handlers

import (
	"config-service/types"
	"config-service/utils/consts"
	"fmt"

	"github.com/gin-gonic/gin"
)

type routerOptions[T types.DocContent] struct {
	dbCollection              string
	path                      string
	nameQueryParam            string
	queryConfig               *queryParamsConfig
	serveGet                  bool //default true, when false, GET will not be served
	servePost                 bool //default true, when false, POST will not be served
	servePut                  bool //default true, when false, PUT will not be served
	serveDelete               bool //default true, when false, DELETE will not be served
	serveGetIncludeGlobalDocs bool //default false, when true, in GET all the response will include global documents (with customers[""])
	serveDeleteByName         bool //default false, when true, DELETE will check for name param and will delete the document by name
	validatePostUniqueName    bool //default true, when true, POST will validate that the name is unique
	validatePutGUID           bool //default true, when true, PUT will validate GUID existence in body or path
	putValidators             []Validator[T]
	postValidators            []Validator[T]
}

func AddRoutes[T types.DocContent](g *gin.Engine, options ...RouterOption[T]) *gin.RouterGroup {
	opts := newRouterOptions[T]()
	opts.apply(options)
	if err := opts.validate(); err != nil {
		panic(err)
	}
	routerGroup := g.Group(opts.path)
	routerGroup.Use(DBContextMiddleware(opts.dbCollection))

	if opts.serveGet {
		routerGroup.GET("", HandleGetByQueryOrAll[T](opts.nameQueryParam, opts.queryConfig, opts.serveGetIncludeGlobalDocs))
		routerGroup.GET("/:"+consts.GUIDField, HandleGetDocWithGUIDInPath[T])
	}
	if opts.servePost {
		postValidators := []Validator[T]{}
		if opts.validatePostUniqueName {
			postValidators = append(postValidators, ValidateUniqueValues(NameKeyGetter[T]))
		}
		postValidators = append(postValidators, opts.postValidators...)
		routerGroup.POST("", HandlePostDocWithValidation(postValidators...)...)
	}
	if opts.servePut {
		putValidators := []Validator[T]{}
		if opts.validatePutGUID {
			putValidators = append(putValidators, ValidateGUIDExistence[T])
		}
		putValidators = append(putValidators, opts.putValidators...)
		routerGroup.PUT("", HandlePutDocWithValidation(putValidators...)...)
		routerGroup.PUT("/:"+consts.GUIDField, HandlePutDocWithValidation(putValidators...)...)
	}
	if opts.serveDelete {
		if opts.serveDeleteByName {
			routerGroup.DELETE("", HandleDeleteDocByName[T](opts.nameQueryParam))
		}
		routerGroup.DELETE("/:"+consts.GUIDField, HandleDeleteDoc[T])
	}
	return routerGroup
}

// Common router config for policies
func AddPolicyRoutes[T types.DocContent](g *gin.Engine, path, dbCollection string, paramConf *queryParamsConfig) *gin.RouterGroup {
	return AddRoutes(g, WithPath[T](path),
		WithDBCollection[T](dbCollection),
		WithNameQuery[T](consts.PolicyNameParam),
		WithQueryConfig[T](paramConf),
		WithIncludeGlobalDocs[T](true),
		WithDeleteByName[T](true),
		WithValidatePostUniqueName[T](true),
		WithValidatePutGUID[T](true),
	)
}

//Options for routes

func newRouterOptions[T types.DocContent]() *routerOptions[T] {
	return &routerOptions[T]{
		serveGet:                  true,
		servePost:                 true,
		servePut:                  true,
		serveDelete:               true,
		serveGetIncludeGlobalDocs: false,
		serveDeleteByName:         false,
		validatePostUniqueName:    true,
		validatePutGUID:           true,
	}

}

func (opts *routerOptions[T]) apply(options []RouterOption[T]) {
	for _, option := range options {
		option(opts)
	}
}

func (opts *routerOptions[T]) validate() error {
	if opts.dbCollection == "" || opts.path == "" {
		return fmt.Errorf("dbCollection and path must be set")
	}
	if opts.serveGetIncludeGlobalDocs && !opts.serveGet {
		return fmt.Errorf("serveGetIncludeGlobalDocs can only be true when serveGet is true")
	}
	if opts.serveDeleteByName && !opts.serveDelete {
		return fmt.Errorf("serveDeleteByName can only be true when serveDelete is true")
	}
	return nil
}

type RouterOption[T types.DocContent] func(*routerOptions[T])

func WithDBCollection[T types.DocContent](dbCollection string) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.dbCollection = dbCollection
	}
}

func WithPath[T types.DocContent](path string) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.path = path
	}
}

func WithServeGet[T types.DocContent](serveGet bool) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.serveGet = serveGet
	}
}

func WithServePost[T types.DocContent](servePost bool) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.servePost = servePost
	}
}

func WithServePut[T types.DocContent](servePut bool) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.servePut = servePut
	}
}

func WithServeDelete[T types.DocContent](serveDelete bool) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.serveDelete = serveDelete
	}
}

func WithIncludeGlobalDocs[T types.DocContent](serveGetIncludeGlobalDocs bool) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.serveGetIncludeGlobalDocs = serveGetIncludeGlobalDocs
	}
}

func WithDeleteByName[T types.DocContent](serveDeleteByName bool) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.serveDeleteByName = serveDeleteByName
	}
}

func WithNameQuery[T types.DocContent](nameQueryParam string) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.nameQueryParam = nameQueryParam
	}
}

func WithQueryConfig[T types.DocContent](queryConfig *queryParamsConfig) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.queryConfig = queryConfig
	}
}

func WithPutValidators[T types.DocContent](validators ...Validator[T]) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.putValidators = validators
	}
}

func WithPostValidators[T types.DocContent](validators ...Validator[T]) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.postValidators = validators
	}
}

func WithValidatePostUniqueName[T types.DocContent](validatePostUniqueName bool) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.validatePostUniqueName = validatePostUniqueName
	}
}

func WithValidatePutGUID[T types.DocContent](validatePutGUID bool) RouterOption[T] {
	return func(opts *routerOptions[T]) {
		opts.validatePutGUID = validatePutGUID
	}
}
