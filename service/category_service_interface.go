package service

import (
	"walk_backend/dto"
	"walk_backend/model"
)

type CategoryServiceInteface interface {
	ListCategories() (model.CategoryList, error)
	Create(dto *dto.Category) (model.ID, error)
	Update(dto *dto.Category) error
	Delete(id model.ID) error
	Find(id model.ID) (*model.Category, error)
}