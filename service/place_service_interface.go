package service

import (
	"walk_backend/dto"
	"walk_backend/model"
)

type PlaceServiceInteface interface {
	ListPlaces() (model.PlaceList, error)
	Create(dto *dto.Place) (model.ID, error)
	Update(dto *dto.Place) error
	Delete(id model.ID) error
	Find(id model.ID) (*model.Place, error)
	Search(search string) (model.PlaceList, error)
	ListCategories() (model.CategoryList, error)
	FindCategory(id model.ID) (*model.Category, error)
}