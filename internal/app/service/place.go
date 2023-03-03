package service

import (
	"context"
	"time"

	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/pkg/cache"

	"github.com/gosimple/slug"
)

const (
	listPlacesCacheKey            string        = "list-places"
	listPlacesCacheDuration       time.Duration = 5 * time.Minute
	searchListPlacesCacheKey      string        = "search-list-places"
	searchListPlacesCacheDuration time.Duration = 5 * time.Minute
)

// PlaceRepositoryInterface ...
type PlaceRepositoryInterface interface {
	Find(ctx context.Context, id model.ID) (*model.Place, error)
	FindAll(ctx context.Context) (model.PlaceList, error)
	Create(ctx context.Context, m *model.Place) (model.ID, error)
	Update(ctx context.Context, m *model.Place) error
	Delete(ctx context.Context, id model.ID) error
	Search(ctx context.Context, search string) (model.PlaceList, error)
}

// PlaceCategoryRepositoryInterface ...
type PlaceCategoryRepositoryInterface interface {
	Find(ctx context.Context, id model.ID) (*model.Category, error)
	FindAll(ctx context.Context) (model.CategoryList, error)
}

// PlaceQueueRepositoryInterface ...
type PlaceQueueRepositoryInterface interface {
	PublishReIndex(id model.ID) error
}

// PlaceCacheRepositoryInterface ...
type PlaceCacheRepositoryInterface interface {
	Get(ctx context.Context, key string) (model.PlaceList, error)
	Set(ctx context.Context, key string, value model.PlaceList, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
}

// DefaultPlaceService ...
type DefaultPlaceService struct {
	placeRepo    PlaceRepositoryInterface
	categoryRepo PlaceCategoryRepositoryInterface
	placeQueue   PlaceQueueRepositoryInterface
	placeCache   PlaceCacheRepositoryInterface
	keyBuilder   cache.KeyBuilderInterface
}

// NewDefaultPlaceService create new default place service
func NewDefaultPlaceService(
	placeRepo PlaceRepositoryInterface,
	categoryRepo PlaceCategoryRepositoryInterface,
	placeQueue PlaceQueueRepositoryInterface,
	placeCache PlaceCacheRepositoryInterface,
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
func (s *DefaultPlaceService) ListPlaces(ctx context.Context) (model.PlaceList, error) {

	places, err := s.placeCache.Get(ctx, listPlacesCacheKey)
	if err != nil {
		return nil, err
	} else if places != nil {
		return places, nil
	}

	places, err = s.placeRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	if err = s.placeCache.Set(ctx, listPlacesCacheKey, places, listPlacesCacheDuration); err != nil {
		return nil, err
	}

	return places, nil
}

// Create ...
func (s *DefaultPlaceService) Create(ctx context.Context, d *dto.Place) (model.ID, error) {

	m, err := s.makeModelFromPlaceDTO(ctx, d)
	if err != nil {
		return model.NilID, err
	}
	m.CreatedAt = time.Now()

	id, err := s.placeRepo.Create(ctx, m)
	if err != nil {
		return model.NilID, err
	}

	if err := s.placeCache.Del(ctx, listPlacesCacheKey); err != nil {
		return model.NilID, err
	}

	if err := s.placeQueue.PublishReIndex(m.ID); err != nil {
		return model.NilID, err
	}

	return id, nil
}

// Update ...
func (s *DefaultPlaceService) Update(ctx context.Context, d *dto.Place) error {

	m, err := s.makeModelFromPlaceDTO(ctx, d)
	if err != nil {
		return err
	}
	m.UpdatedAt = time.Now()

	if err := s.placeRepo.Update(ctx, m); err != nil {
		return err
	}

	if err := s.placeCache.Del(ctx, listPlacesCacheKey); err != nil {
		return err
	}

	if err := s.placeQueue.PublishReIndex(m.ID); err != nil {
		return err
	}

	return nil
}

// Delete ...
func (s *DefaultPlaceService) Delete(ctx context.Context, id model.ID) error {

	if err := s.placeRepo.Delete(ctx, id); err != nil {
		return err
	}

	if err := s.placeCache.Del(ctx, listPlacesCacheKey); err != nil {
		return err
	}

	if err := s.placeQueue.PublishReIndex(id); err != nil {
		return err
	}

	return nil
}

// Find ...
func (s *DefaultPlaceService) Find(ctx context.Context, id model.ID) (*model.Place, error) {
	return s.placeRepo.Find(ctx, id)
}

// Search ...
func (s *DefaultPlaceService) Search(ctx context.Context, search string) (model.PlaceList, error) {

	key := s.keyBuilder.NewKey()
	key.Add(searchListPlacesCacheKey)
	if err := key.AddHashed(search); err != nil {
		return nil, err
	}
	cacheKey := key.String()

	places, err := s.placeCache.Get(ctx, cacheKey)
	if err != nil {
		return nil, err
	} else if places != nil {
		return places, nil
	}

	places, err = s.placeRepo.Search(ctx, search)
	if err != nil {
		return nil, err
	}

	if err = s.placeCache.Set(ctx, cacheKey, places, searchListPlacesCacheDuration); err != nil {
		return nil, err
	}

	return places, nil
}

// ListCategories ...
func (s *DefaultPlaceService) ListCategories(ctx context.Context) (model.CategoryList, error) {
	return s.categoryRepo.FindAll(ctx)
}

// FindCategory ...
func (s *DefaultPlaceService) FindCategory(ctx context.Context, id model.ID) (*model.Category, error) {
	return s.categoryRepo.Find(ctx, id)
}

func (s *DefaultPlaceService) makeModelFromPlaceDTO(ctx context.Context, d *dto.Place) (*model.Place, error) {

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

	if _, err := s.categoryRepo.Find(ctx, categoryID); err != nil {
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
