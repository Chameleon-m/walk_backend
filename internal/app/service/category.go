package service

import (
	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/app/repository"
)

// DefaultCategoryService ...
type DefaultCategoryService struct {
	categoryRepo repository.CategoryRepositoryInterface
}

var _ CategoryServiceInteface = (*DefaultCategoryService)(nil)

// NewDefaultCategoryService create new default category service
func NewDefaultCategoryService(categoryRepo repository.CategoryRepositoryInterface) *DefaultCategoryService {
	return &DefaultCategoryService{
		categoryRepo: categoryRepo,
	}
}

// ListCategories ...
func (s *DefaultCategoryService) ListCategories() (model.CategoryList, error) {
	return s.categoryRepo.FindAll()
}

// Create ...
func (s *DefaultCategoryService) Create(d *dto.Category) (model.ID, error) {

	m, err := s.makeModelFromCategoryDTO(d)
	if err != nil {
		return model.NilID, err
	}

	return s.categoryRepo.Create(m)
}

// Update ...
func (s *DefaultCategoryService) Update(d *dto.Category) error {

	m, err := s.makeModelFromCategoryDTO(d)
	if err != nil {
		return err
	}

	return s.categoryRepo.Update(m)
}

// Delete ...
func (s *DefaultCategoryService) Delete(id model.ID) error {
	return s.categoryRepo.Delete(id)
}

// Find ...
func (s *DefaultCategoryService) Find(id model.ID) (*model.Category, error) {
	return s.categoryRepo.Find(id)
}

func (s *DefaultCategoryService) makeModelFromCategoryDTO(d *dto.Category) (*model.Category, error) {

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

	m, err := model.NewCategoryModel(
		id,
		d.Name,
		d.Order,
	)
	if err != nil {
		return nil, err
	}

	return m, nil
}
