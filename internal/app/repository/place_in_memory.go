package repository

import (
	"sync"
	"time"
	"walk_backend/internal/app/model"
)

type PlaceInMemoryRepository struct {
	lock   sync.RWMutex
	Places model.PlaceList
}

var _ PlaceRepositoryInterface = (*PlaceInMemoryRepository)(nil)

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

func (r *PlaceInMemoryRepository) Create(m *model.Place) (model.ID, error) {
	if m.ID.IsNil() {
		id, err := model.NewID()
		if err != nil {
			return model.NilID, err
		}
		m.ID = id
	}
	m.CreatedAt = time.Now()

	r.lock.Lock()
	r.Places = append(r.Places, m)
	r.lock.Unlock()

	return m.ID, nil
}

func (r *PlaceInMemoryRepository) Update(m *model.Place) error {

	r.lock.Lock()
	defer r.lock.Unlock()

	for k := 0; k < len(r.Places); k++ {
		if r.Places[k].ID == m.ID {
			m.UpdatedAt = time.Now()
			r.Places[k] = m
			return nil
		}
	}

	return model.ErrModelNotFound
}

func (r *PlaceInMemoryRepository) Delete(id model.ID) error {

	r.lock.Lock()
	defer r.lock.Unlock()

	for k := 0; k < len(r.Places); k++ {
		if r.Places[k].ID == id {
			r.Places = append(r.Places[:k], r.Places[k+1:]...)
			return nil
		}
	}

	return model.ErrModelNotFound
}

func (r *PlaceInMemoryRepository) Search(search string) (model.PlaceList, error) {
	panic("need implement")
}
