package handlers

import (
	"errors"
	"net/http"

	"walk_backend/cmd/api/presenter"
	"walk_backend/dto"
	"walk_backend/model"
	"walk_backend/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

type CategoriesHandler struct {
	service   service.CategoryServiceInteface
	ctx       context.Context
	presenter presenter.CategoryPresenterInteface
}

func NewCategoriesHandler(ctx context.Context, service service.CategoryServiceInteface, presenter presenter.CategoryPresenterInteface) *CategoriesHandler {
	return &CategoriesHandler{
		service:   service,
		ctx:       ctx,
		presenter: presenter,
	}
}

// swagger:operation GET /categories categories ListCategories
// Returns list of categories
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Successful operation
func (handler *CategoriesHandler) ListCategoriesHandler(c *gin.Context) {

	categoryList, err := handler.service.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data := handler.presenter.MakeList(categoryList)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// swagger:operation POST /categories categories newCategory
// Create a new category
// ---
// produces:
// - application/json
// responses:
//
//	'201':
//	    description: Successful operation
//	'400':
//	    description: Invalid input
func (handler *CategoriesHandler) NewCategoryHandler(c *gin.Context) {

	dto := dto.NewCategoryDTO()
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

	c.Header("Location", makeUrl(c.Request, "/v1/categories/"+id.String()))
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

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
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid category ID
func (handler *CategoriesHandler) UpdateCategryHandler(c *gin.Context) {

	dto := dto.NewCategoryDTO()
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
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid category ID
func (handler *CategoriesHandler) DeleteCategoryHandler(c *gin.Context) {
	id := c.Param("id")
	categoryID, err := model.StringToID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := handler.service.Delete(categoryID); err != nil {
		if errors.Is(err, model.ErrModelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

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
//	    description: Successful operation
//	'400':
//	    description: Invalid input
//	'404':
//	    description: Invalid category ID
func (handler *CategoriesHandler) GetOneCategoryHandler(c *gin.Context) {
	id := c.Param("id")
	categoryID, err := model.StringToID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := handler.service.Find(categoryID)
	if err != nil {
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

func (handler *CategoriesHandler) MakeHandlers(router *gin.RouterGroup, routerAuth *gin.RouterGroup) {

	router.GET("/categories", handler.ListCategoriesHandler)
	router.GET("/categories/:id", handler.GetOneCategoryHandler)

	routerAuth.POST("/categories", handler.NewCategoryHandler)
	routerAuth.PUT("/categories/:id", handler.UpdateCategryHandler)
	routerAuth.DELETE("/categories/:id", handler.DeleteCategoryHandler)
}
