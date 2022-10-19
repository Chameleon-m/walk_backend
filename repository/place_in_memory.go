package repository

import (
	"sync"
	"time"
	"walk_backend/model"
)

type PlaceInMemoryRepository struct {
	lock   sync.RWMutex
	Places model.PlaceList
}

func NewPlaceInMemoryRepository() *PlaceInMemoryRepository {
	var places = model.PlaceList{}
	return &PlaceInMemoryRepository{
		Places: places,
	}
}

func (r *PlaceInMemoryRepository) Find(id model.ID) (*model.Place, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for k := 0; k < len(r.Places); k++ {
		if r.Places[k].ID == id {
			return r.Places[k], nil
		}
	}

	return nil, model.ErrModelNotFound
}

func (r *PlaceInMemoryRepository) FindAll() (model.PlaceList, error) {
	return r.Places, nil
}

func (r *PlaceInMemoryRepository) FindBy(criteria model.Criteria, orderBy []string, limit int64, offset int64) (model.PlaceList, error) {
	panic("need implement")
}

func (r *PlaceInMemoryRepository) FindOneBy(criteria model.Criteria) (*model.Place, error) {
	panic("need implement")
}

func (r *PlaceInMemoryRepository) Count(criteria model.Criteria) (int, error) {
	return len(r.Places), nil
}

func (r *PlaceInMemoryRepository) Create(place model.Place) (model.ID, error) {
	if place.ID.IsNil() {
		id, err := model.NewID()
		if err != nil {
			return model.NilID, err
		}
		place.ID = id
	}
	place.CreatedAt = time.Now()

	r.lock.Lock()
	r.Places = append(r.Places, &place)
	r.lock.Unlock()

	return place.ID, nil
}

func (r *PlaceInMemoryRepository) Update(place model.Place) error {

	r.lock.Lock()
	defer r.lock.Unlock()

	for k := 0; k < len(r.Places); k++ {
		if r.Places[k].ID == place.ID {
			place.UpdatedAt = time.Now()
			r.Places[k] = &place
			return nil
		}
	}

	return model.ErrModelNotFound
}

func (r *PlaceInMemoryRepository) Delete(place model.Place) error {

	r.lock.Lock()
	defer r.lock.Unlock()

	for k := 0; k < len(r.Places); k++ {
		if r.Places[k].ID == place.ID {
			r.Places = append(r.Places[:k], r.Places[k+1:]...)
			return nil
		}
	}

	return model.ErrModelNotFound
}

func (r *PlaceInMemoryRepository) Search(search string) (model.PlaceList, error) {
	panic("need implement")
}
