package repository

import (
	"walk_backend/model"
)

type PlaceRepositoryInterface interface {
	Find(id model.ID) (*model.Place, error)
	FindAll() (model.PlaceList, error)
	FindBy(criteria model.Criteria, orderBy []string, limit int64, offset int64) (model.PlaceList, error)
	FindOneBy(criteria model.Criteria) (*model.Place, error)
	Count(criteria model.Criteria) (int, error)
	Create(m *model.Place) (model.ID, error)
	Update(m *model.Place) error
	Delete(id model.ID) error
	Search(search string) (model.PlaceList, error)
}
