package service

import (
	"time"
	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/app/repository"

	"github.com/gosimple/slug"
)

type DefaultPlaceService struct {
	placeRepo    repository.PlaceRepositoryInterface
	categoryRepo repository.CategoryRepositoryInterface
	placeQueue   repository.PlaceQueueRepositoryInterface
}

func NewDefaultPlaceService(
	placeRepo repository.PlaceRepositoryInterface,
	categoryRepo repository.CategoryRepositoryInterface,
	placeQueue repository.PlaceQueueRepositoryInterface,
) *DefaultPlaceService {
	return &DefaultPlaceService{
		placeRepo:    placeRepo,
		categoryRepo: categoryRepo,
		placeQueue:   placeQueue,
	}
}

func (s *DefaultPlaceService) ListPlaces() (model.PlaceList, error) {
	return s.placeRepo.FindAll()
}

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

	if err := s.placeQueue.PublishReIndex(m.ID); err != nil {
		return model.NilID, err
	}

	return id, nil
}

func (s *DefaultPlaceService) Update(d *dto.Place) error {

	m, err := s.makeModelFromPlaceDTO(d)
	if err != nil {
		return err
	}
	m.UpdatedAt = time.Now()

	if err := s.placeRepo.Update(m); err != nil {
		return err
	}

	if err := s.placeQueue.PublishReIndex(m.ID); err != nil {
		return err
	}

	return nil
}

func (s *DefaultPlaceService) Delete(id model.ID) error {

	if err := s.placeRepo.Delete(id); err != nil {
		return err
	}

	if err := s.placeQueue.PublishReIndex(id); err != nil {
		return err
	}

	return nil
}

func (s *DefaultPlaceService) Find(id model.ID) (*model.Place, error) {
	return s.placeRepo.Find(id)
}

func (s *DefaultPlaceService) Search(search string) (model.PlaceList, error) {
	return s.placeRepo.Search(search)
}

func (s *DefaultPlaceService) ListCategories() (model.CategoryList, error) {
	return s.categoryRepo.FindAll()
}

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
