package repository

import (
	"walk_backend/internal/app/model"
)

// PlaceRepositoryInterface ...
type PlaceRepositoryInterface interface {
	Find(id model.ID) (*model.Place, error)
	FindAll() (model.PlaceList, error)
	Create(m *model.Place) (model.ID, error)
	Update(m *model.Place) error
	Delete(id model.ID) error
	Search(search string) (model.PlaceList, error)
}
