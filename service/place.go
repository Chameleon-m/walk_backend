package service

import (
	"time"
	"walk_backend/dto"
	"walk_backend/model"
	"walk_backend/repository"

	"github.com/gosimple/slug"
)

type DefaultPlaceService struct {
	placeRepo    repository.PlaceRepositoryInterface
	categoryRepo repository.CategoryRepositoryInterface
}

func NewDefaultPlaceService(placeRepo repository.PlaceRepositoryInterface, categoryRepo repository.CategoryRepositoryInterface) *DefaultPlaceService {
	return &DefaultPlaceService{
		placeRepo:    placeRepo,
		categoryRepo: categoryRepo,
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

	return s.placeRepo.Create(m)
}

func (s *DefaultPlaceService) Update(d *dto.Place) error {

	m, err := s.makeModelFromPlaceDTO(d)
	if err != nil {
		return err
	}
	m.UpdatedAt = time.Now()

	return s.placeRepo.Update(m)
}

func (s *DefaultPlaceService) Delete(id model.ID) error {
	return s.placeRepo.Delete(id)
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
