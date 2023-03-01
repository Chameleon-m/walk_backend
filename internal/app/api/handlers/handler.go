package handlers

import (
	"walk_backend/internal/pkg/httpserver"

	"github.com/gin-gonic/gin"
)

type CategoriesHandlerInterface interface {
	HandlerWithAuthInterface
	ListCategoriesHandler(c *gin.Context)
	NewCategoryHandler(c *gin.Context)
	UpdateCategryHandler(c *gin.Context)
	DeleteCategoryHandler(c *gin.Context)
	GetOneCategoryHandler(c *gin.Context)
}

type AuthHandlerInterface interface {
	HandlerInterface
	SignUpHandler(c *gin.Context)
	SignInHandler(c *gin.Context)
	RefreshHandler(c *gin.Context)
	SignOutHandler(c *gin.Context)
}

type PlacesHandlerInterface interface {
	HandlerWithAuthInterface
	HandlerRequestValidationInterface
	ListPlacesHandler(c *gin.Context)
	NewPlaceHandler(c *gin.Context)
	UpdatePlaceHandler(c *gin.Context)
	DeletePlaceHandler(c *gin.Context)
	GetOnePlaceHandler(c *gin.Context)
	SearchPlacesHandler(c *gin.Context)
}

// HandlerInterface ...
type HandlerInterface interface {
	MakeHandlers(router *gin.RouterGroup)
}

// HandlerWithAuthInterface ...
type HandlerWithAuthInterface interface {
	MakeHandlers(router *gin.RouterGroup, routerAuth *gin.RouterGroup)
}

// HandlerValidateInterface ...
type HandlerRequestValidationInterface interface {
	MakeRequestValidation()
}

// HandlersInterface ...
type HandlersInterface interface {
	Make()
	GetAuthHandler() AuthHandlerInterface
	SetAuthHandler(handler AuthHandlerInterface)
	GetCategoriesHandler() CategoriesHandlerInterface
	SetCategoriesHandler(handler CategoriesHandlerInterface)
	GetPlacesHandler() PlacesHandlerInterface
	SetPlacesHandler(handler PlacesHandlerInterface)
}

type handlers struct {
	app               httpserver.ServerInterface
	router            *gin.RouterGroup
	routerAuth        *gin.RouterGroup
	authHandler       AuthHandlerInterface
	placesHandler     PlacesHandlerInterface
	categoriesHandler CategoriesHandlerInterface
}

// New ...
func New(app httpserver.ServerInterface, router *gin.RouterGroup, routerAuth *gin.RouterGroup) *handlers {
	return &handlers{
		app:        app,
		router:     router,
		routerAuth: routerAuth,
	}
}

// GetAuthHandler ...
func (h *handlers) GetAuthHandler() AuthHandlerInterface {
	return h.authHandler
}

// SetAuthHandler ...
func (h *handlers) SetAuthHandler(handler AuthHandlerInterface) {
	h.authHandler = handler
}

// GetCategoriesHandler ...
func (h *handlers) GetCategoriesHandler() CategoriesHandlerInterface {
	return h.categoriesHandler
}

// SetCategoriesHandler ...
func (h *handlers) SetCategoriesHandler(handler CategoriesHandlerInterface) {
	h.categoriesHandler = handler
}

// GetPlacesHandler ...
func (h *handlers) GetPlacesHandler() PlacesHandlerInterface {
	return h.placesHandler
}

// SetPlacesHandler ...
func (h *handlers) SetPlacesHandler(handler PlacesHandlerInterface) {
	h.placesHandler = handler
}

func (h *handlers) Make() {
	h.GetAuthHandler().MakeHandlers(h.router)
	h.GetPlacesHandler().MakeHandlers(h.router, h.routerAuth)
	h.GetPlacesHandler().MakeRequestValidation()
	h.GetCategoriesHandler().MakeHandlers(h.router, h.routerAuth)
}
