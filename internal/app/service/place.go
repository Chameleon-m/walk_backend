package service

import (
	"time"

	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/app/repository"
	"walk_backend/internal/pkg/cache"

	"github.com/gosimple/slug"
)

const (
	listPlacesCacheKey            string        = "list-places"
	listPlacesCacheDuration       time.Duration = 5 * time.Minute
	searchListPlacesCacheKey      string        = "search-list-places"
	searchListPlacesCacheDuration time.Duration = 5 * time.Minute
)

// DefaultPlaceService ...
type DefaultPlaceService struct {
	placeRepo    repository.PlaceRepositoryInterface
	categoryRepo repository.CategoryRepositoryInterface
	placeQueue   repository.PlaceQueueRepositoryInterface
	placeCache   repository.PlaceCacheRepositoryInterface
	keyBuilder   cache.KeyBuilderInterface
}

var _ PlaceServiceInteface = (*DefaultPlaceService)(nil)

// NewDefaultPlaceService create new default place service
func NewDefaultPlaceService(
	placeRepo repository.PlaceRepositoryInterface,
	categoryRepo repository.CategoryRepositoryInterface,
	placeQueue repository.PlaceQueueRepositoryInterface,
	placeCache repository.PlaceCacheRepositoryInterface,
	keyBuilder cache.KeyBuilderInterface,
) *DefaultPlaceService {
	return &DefaultPlaceService{
		placeRepo:    placeRepo,
		categoryRepo: categoryRepo,
		placeQueue:   placeQueue,
		placeCache:   placeCache,
		keyBuilder:   keyBuilder,
	}
}

// ListPlaces ...
func (s *DefaultPlaceService) ListPlaces() (model.PlaceList, error) {

	places, err := s.placeCache.Get(listPlacesCacheKey)
	if err != nil {
		return nil, err
	} else if places != nil {
		return places, nil
	}

	places, err = s.placeRepo.FindAll()
	if err != nil {
		return nil, err
	}

	if err = s.placeCache.Set(listPlacesCacheKey, places, listPlacesCacheDuration); err != nil {
		return nil, err
	}

	return places, nil
}

// Create ...
func (s *DefaultPlaceService) Create(d *dto.Place) (model.ID, error) {

	m, err := s.makeModelFromPlaceDTO(d)
	if err != nil {
		return model.NilID, err
	}
	m.CreatedAt = time.Now()

	id, err := s.placeRepo.Create(m)
	if err != nil {
		return model.NilID, err
	}

	if err := s.placeCache.Del(listPlacesCacheKey); err != nil {
		return model.NilID, err
	}

	if err := s.placeQueue.PublishReIndex(m.ID); err != nil {
		return model.NilID, err
	}

	return id, nil
}

// Update ...
func (s *DefaultPlaceService) Update(d *dto.Place) error {

	m, err := s.makeModelFromPlaceDTO(d)
	if err != nil {
		return err
	}
	m.UpdatedAt = time.Now()

	if err := s.placeRepo.Update(m); err != nil {
		return err
	}

	if err := s.placeCache.Del(listPlacesCacheKey); err != nil {
		return err
	}

	if err := s.placeQueue.PublishReIndex(m.ID); err != nil {
		return err
	}

	return nil
}

// Delete ...
func (s *DefaultPlaceService) Delete(id model.ID) error {

	if err := s.placeRepo.Delete(id); err != nil {
		return err
	}

	if err := s.placeCache.Del(listPlacesCacheKey); err != nil {
		return err
	}

	if err := s.placeQueue.PublishReIndex(id); err != nil {
		return err
	}

	return nil
}

// Find ...
func (s *DefaultPlaceService) Find(id model.ID) (*model.Place, error) {
	return s.placeRepo.Find(id)
}

// Search ...
func (s *DefaultPlaceService) Search(search string) (model.PlaceList, error) {

	key := s.keyBuilder.NewKey()
	key.Add(searchListPlacesCacheKey)
	if err := key.AddHashed(search); err != nil {
		return nil, err
	}
	cacheKey := key.String()

	places, err := s.placeCache.Get(cacheKey)
	if err != nil {
		return nil, err
	} else if places != nil {
		return places, nil
	}

	places, err = s.placeRepo.Search(search)
	if err != nil {
		return nil, err
	}

	if err = s.placeCache.Set(cacheKey, places, searchListPlacesCacheDuration); err != nil {
		return nil, err
	}

	return places, nil
}

// ListCategories ...
func (s *DefaultPlaceService) ListCategories() (model.CategoryList, error) {
	return s.categoryRepo.FindAll()
}

// FindCategory ...
func (s *DefaultPlaceService) FindCategory(id model.ID) (*model.Category, error) {
	return s.categoryRepo.Find(id)
}

func (s *DefaultPlaceService) makeModelFromPlaceDTO(d *dto.Place) (*model.Place, error) {

	var id model.ID
	var err error
	if d.ID != "" {
		id, err = model.StringToID(d.ID)
	} else {
		id, err = model.NewID()
	}

	if err != nil {
		return nil, err
	}

	categoryID, err := model.StringToID(d.Category)
	if err != nil {
		return nil, err
	}

	if _, err := s.categoryRepo.Find(categoryID); err != nil {
		return nil, err
	}

	m, err := model.NewPlaceModel(
		id,
		d.Name,
		slug.Make(d.Name),
		d.Description,
		categoryID,
		d.Tags,
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}
