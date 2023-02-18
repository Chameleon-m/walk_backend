package handlers

import (
	"walk_backend/internal/pkg/httpserver"

	"github.com/gin-gonic/gin"
)

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
	GetAuthHandler() AuthHandlerInterface
	SetAuthHandler(handler AuthHandlerInterface)
	GetCategoriesHandler() CategoriesHandlerInterface
	SetCategoriesHandler(handler CategoriesHandlerInterface)
	GetPlacesHandler() PlacesHandlerInterface
	SetPlacesHandler(handler PlacesHandlerInterface)
}

type handlers struct {
	app               httpserver.ServerInterface
	authHandler       AuthHandlerInterface
	placesHandler     PlacesHandlerInterface
	categoriesHandler CategoriesHandlerInterface
}

// New ...
func New(app httpserver.ServerInterface) *handlers {
	return &handlers{
		app: app,
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

// func (h *handlers) Make() {
// 	h.GetAuthHandler().MakeHandlers(apiV1)
// 	h.GetPlacesHandler().MakeHandlers(apiV1, apiV1auth)
// 	h.GetPlacesHandler().MakeRequestValidation()
// 	h.GetCategoriesHandler().MakeHandlers(apiV1, apiV1auth)
// }
