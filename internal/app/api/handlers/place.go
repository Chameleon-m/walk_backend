package handlers

import (
	"errors"
	"net/http"

	"walk_backend/internal/app/api/presenter"
	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"golang.org/x/net/context"
)

// PlacesHandler ...
type PlacesHandler struct {
	service   service.PlaceServiceInteface
	ctx       context.Context
	presenter presenter.PlacePresenterInteface
}

// NewPlacesHandler ...
func NewPlacesHandler(ctx context.Context, service service.PlaceServiceInteface, presenter presenter.PlacePresenterInteface) *PlacesHandler {
	return &PlacesHandler{
		service:   service,
		ctx:       ctx,
		presenter: presenter,
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

	placeList, err := handler.service.ListPlaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	categoryList, err := handler.service.ListCategories()
	if err != nil {
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

	id, err := handler.service.Create(dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Location", makeURL(c.Request, "/api/v1/places/"+id.String()))
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

	if err := handler.service.Update(dto); err != nil {
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

	if err := handler.service.Delete(placeID); err != nil {
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

	place, err := handler.service.Find(placeID)
	if err != nil {
		if errors.Is(err, model.ErrModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	category, err := handler.service.FindCategory(place.Category)
	if err != nil {
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
	placeList, err := handler.service.Search(search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	categoryList, err := handler.service.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else if len(categoryList) == 0 {
		c.JSON(http.StatusFailedDependency, gin.H{"error": "Categories not found"})
		return
	}

	data := handler.presenter.MakeList(placeList, categoryList)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// MakeHandlers make places routes
func (handler *PlacesHandler) MakeHandlers(router *gin.RouterGroup, routerAuth *gin.RouterGroup) {

	router.GET("/places", handler.ListPlacesHandler)
	router.GET("/places/:id", handler.GetOnePlaceHandler)
	router.GET("/places/search", handler.SearchPlacesHandler)

	routerAuth.POST("/places", handler.NewPlaceHandler)
	routerAuth.PUT("/places/:id", handler.UpdatePlaceHandler)
	routerAuth.DELETE("/places/:id", handler.DeletePlaceHandler)
}

// MakeRequestValidation make request validation
func (handler *PlacesHandler) MakeRequestValidation() {

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterStructValidation(dto.ValidatePlaceDTO, dto.NewPlaceDTO())
	}
}
