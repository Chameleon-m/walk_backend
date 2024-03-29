package place

import (
	"errors"
	"net/http"

	"walk_backend/internal/app/api/presenter"
	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"golang.org/x/net/context"
)

// ServiceInterface ...
type ServiceInterface interface {
	ListPlaces(ctx context.Context) (model.PlaceList, error)
	Create(ctx context.Context, dto *dto.Place) (model.ID, error)
	Update(ctx context.Context, dto *dto.Place) error
	Delete(ctx context.Context, id model.ID) error
	Find(ctx context.Context, id model.ID) (*model.Place, error)
	Search(ctx context.Context, search string) (model.PlaceList, error)
	ListCategories(ctx context.Context) (model.CategoryList, error)
	FindCategory(ctx context.Context, id model.ID) (*model.Category, error)
}

// PresenterInterface ...
type PresenterInterface interface {
	Make(m *model.Place, c *model.Category) *presenter.Place
	MakeList(mList model.PlaceList, cList model.CategoryList) []*presenter.Place
}

// PlacesHandler ...
type PlacesHandler struct {
	ctx        context.Context
	router     *gin.RouterGroup
	routerAuth *gin.RouterGroup
	service    ServiceInterface
	presenter  PresenterInterface
}

// NewHandler ...
func NewHandler(
	ctx context.Context,
	router *gin.RouterGroup,
	routerAuth *gin.RouterGroup,
	service ServiceInterface,
	presenter PresenterInterface,
) *PlacesHandler {
	return &PlacesHandler{
		ctx:        ctx,
		router:     router,
		routerAuth: routerAuth,
		service:    service,
		presenter:  presenter,
	}
}

// ListPlacesHandler ...
//
// swagger:operation GET /places places listPlaces
// Returns list of places
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	  description: Successful operation
func (handler *PlacesHandler) ListPlacesHandler(c *gin.Context) {

	placeList, err := handler.service.ListPlaces(handler.ctx)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	categoryList, err := handler.service.ListCategories(handler.ctx)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if len(categoryList) == 0 {
		c.JSON(http.StatusFailedDependency, gin.H{"error": "Categories not found"})
		return
	}

	data := handler.presenter.MakeList(placeList, categoryList)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// NewPlaceHandler ...
//
// swagger:operation POST /places places newPlace
// Create a new place
// ---
// produces:
// - application/json
// responses:
//
//	'201':
//	  description: Successful operation
//	'400':
//	  description: Invalid input
func (handler *PlacesHandler) NewPlaceHandler(c *gin.Context) {

	dto := dto.NewPlaceDTO()
	if err := c.ShouldBindJSON(dto); err != nil {
		// You will need to assert the error if it's not nil eg. err.(validator.ValidationErrors) to access the array of errors.
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := handler.service.Create(handler.ctx, dto)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Location", util.MakeURL(c.Request, "/api/v1/places/"+id.String()))
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// UpdatePlaceHandler ...
//
// swagger:operation PUT /places/{id} places updatePlace
// Update an existing place
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the place
//     required: true
//     type: string
//
// produces:
// - application/json
// responses:
//
//	'204':
//	  description: Successful operation
//	'400':
//	  description: Invalid input
//	'404':
//	  description: Invalid place ID
func (handler *PlacesHandler) UpdatePlaceHandler(c *gin.Context) {

	dto := dto.NewPlaceDTO()
	dto.ID = c.Param("id")
	if err := c.ShouldBindJSON(dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := handler.service.Update(handler.ctx, dto); err != nil {
		_ = c.Error(err)
		if errors.Is(err, model.ErrModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, model.ErrModelUpdate) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeletePlaceHandler ...
//
// swagger:operation DELETE /places/{id} places deletePlace
// Delete an existing place
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID of the place
//     required: true
//     type: string
//
// responses:
//
//	'204':
//	  description: Successful operation
//	'400':
//	  description: Invalid input
//	'404':
//	  description: Invalid place ID
func (handler *PlacesHandler) DeletePlaceHandler(c *gin.Context) {
	id := c.Param("id")
	placeID, err := model.StringToID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := handler.service.Delete(handler.ctx, placeID); err != nil {
		_ = c.Error(err)
		if errors.Is(err, model.ErrModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetOnePlaceHandler ...
//
// swagger:operation GET /places/{id} places findPlaceByID
// Get one place
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: place ID
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	  description: Successful operation
//	'400':
//	  description: Invalid input
//	'404':
//	  description: Invalid place ID
func (handler *PlacesHandler) GetOnePlaceHandler(c *gin.Context) {
	id := c.Param("id")
	placeID, err := model.StringToID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	place, err := handler.service.Find(handler.ctx, placeID)
	if err != nil {
		_ = c.Error(err)
		if errors.Is(err, model.ErrModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	category, err := handler.service.FindCategory(handler.ctx, place.Category)
	if err != nil {
		_ = c.Error(err)
		if errors.Is(err, model.ErrModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := handler.presenter.Make(place, category)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// SearchPlacesHandler ...
//
// swagger:operation GET /places/search places findPlace
// Search places based on name, description and tags
// ---
// produces:
// - application/json
// parameters:
//   - name: q
//     in: query
//     description: place name, description and tags
//     required: true
//     type: string
//
// responses:
//
//	'200':
//	  description: Successful operation
func (handler *PlacesHandler) SearchPlacesHandler(c *gin.Context) {
	search := c.Query("q")
	placeList, err := handler.service.Search(handler.ctx, search)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	categoryList, err := handler.service.ListCategories(handler.ctx)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if len(categoryList) == 0 {
		c.JSON(http.StatusFailedDependency, gin.H{"error": "Categories not found"})
		return
	}

	data := handler.presenter.MakeList(placeList, categoryList)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Make ...
func (handler *PlacesHandler) Make() {
	handler.MakeRoutes()
	handler.MakeRequestValidation()
}

// MakeHandlers make places routes
func (handler *PlacesHandler) MakeRoutes() {

	handler.router.GET("/places", handler.ListPlacesHandler)
	handler.router.GET("/places/:id", handler.GetOnePlaceHandler)
	handler.router.GET("/places/search", handler.SearchPlacesHandler)

	handler.routerAuth.POST("/places", handler.NewPlaceHandler)
	handler.routerAuth.PUT("/places/:id", handler.UpdatePlaceHandler)
	handler.routerAuth.DELETE("/places/:id", handler.DeletePlaceHandler)
}

// MakeRequestValidation make request validation
func (handler *PlacesHandler) MakeRequestValidation() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterStructValidation(dto.ValidatePlaceDTO, dto.NewPlaceDTO())
	}
}
