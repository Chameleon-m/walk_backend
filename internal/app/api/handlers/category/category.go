package category

import (
	"errors"
	"net/http"

	"walk_backend/internal/app/api/presenter"
	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/pkg/util"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

// ServiceInterface ...
type ServiceInterface interface {
	ListCategories(ctx context.Context) (model.CategoryList, error)
	Create(ctx context.Context, dto *dto.Category) (model.ID, error)
	Update(ctx context.Context, dto *dto.Category) error
	Delete(ctx context.Context, id model.ID) error
	Find(ctx context.Context, id model.ID) (*model.Category, error)
}

type PresenterInterface interface {
	Make(m *model.Category) *presenter.Category
	MakeList(mList model.CategoryList) []*presenter.Category
}

// CategoriesHandler categories handler struct
type CategoriesHandler struct {
	ctx        context.Context
	router     *gin.RouterGroup
	routerAuth *gin.RouterGroup
	service    ServiceInterface
	presenter  PresenterInterface
}

// NewHandler create new categories handler
func NewHandler(
	ctx context.Context,
	router *gin.RouterGroup,
	routerAuth *gin.RouterGroup,
	service ServiceInterface,
	presenter PresenterInterface,
) *CategoriesHandler {
	return &CategoriesHandler{
		ctx:        ctx,
		router:     router,
		routerAuth: routerAuth,
		service:    service,
		presenter:  presenter,
	}
}

// ListCategoriesHandler ...
//
// swagger:operation GET /categories categories ListCategories
// Returns list of categories
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	  description: Successful operation
func (handler *CategoriesHandler) ListCategoriesHandler(c *gin.Context) {

	categoryList, err := handler.service.ListCategories(handler.ctx)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := handler.presenter.MakeList(categoryList)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// NewCategoryHandler ...
//
// swagger:operation POST /categories categories newCategory
// Create a new category
// ---
// produces:
// - application/json
// responses:
//
//	'201':
//	  description: Successful operation
//	'400':
//	  description: Invalid input
func (handler *CategoriesHandler) NewCategoryHandler(c *gin.Context) {

	dto := dto.NewCategoryDTO()
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

	c.Header("Location", util.MakeURL(c.Request, "/api/v1/categories/"+id.String()))
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// UpdateCategryHandler ...
//
// swagger:operation PUT /categories/{id} categories updateCategory
// Update an existing category
// ---
// parameters:
//   - name: id
//     in: path
//     description: ID of the category
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
//	  description: Invalid category ID
func (handler *CategoriesHandler) UpdateCategryHandler(c *gin.Context) {

	dto := dto.NewCategoryDTO()
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

// DeleteCategoryHandler ...
//
// swagger:operation DELETE /categories/{id} categories deleteCategory
// Delete an existing category
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: ID of the category
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
//	  description: Invalid category ID
func (handler *CategoriesHandler) DeleteCategoryHandler(c *gin.Context) {
	id := c.Param("id")
	categoryID, err := model.StringToID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := handler.service.Delete(handler.ctx, categoryID); err != nil {
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

// GetOneCategoryHandler ...
//
// swagger:operation GET /categories/{id} categories findCategoryByID
// Get one category
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: category ID
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
//	  description: Invalid category ID
func (handler *CategoriesHandler) GetOneCategoryHandler(c *gin.Context) {
	id := c.Param("id")
	categoryID, err := model.StringToID(id)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := handler.service.Find(handler.ctx, categoryID)
	if err != nil {
		_ = c.Error(err)
		if errors.Is(err, model.ErrModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := handler.presenter.Make(category)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// Make ...
func (handler *CategoriesHandler) Make() {
	handler.MakeRoutes()
}

// MakeRoutes ...
func (handler *CategoriesHandler) MakeRoutes() {

	handler.router.GET("/categories", handler.ListCategoriesHandler)
	handler.router.GET("/categories/:id", handler.GetOneCategoryHandler)

	handler.routerAuth.POST("/categories", handler.NewCategoryHandler)
	handler.routerAuth.PUT("/categories/:id", handler.UpdateCategryHandler)
	handler.routerAuth.DELETE("/categories/:id", handler.DeleteCategoryHandler)
}
