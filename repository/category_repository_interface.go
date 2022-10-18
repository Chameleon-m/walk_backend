package repository

import (
	"walk_backend/model"
)

type CategoryRepositoryInterface interface {
	Find(id model.ID) (*model.Category, error)
	FindAll() (model.CategoryList, error)
	Create(m *model.Category) (model.ID, error)
	Update(m *model.Category) error
	Delete(id model.ID) error
}
