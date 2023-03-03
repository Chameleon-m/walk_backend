package service

import (
	"context"
	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
)

// CategoryRepositoryInterface ...
type CategoryRepositoryInterface interface {
	Find(ctx context.Context, id model.ID) (*model.Category, error)
	FindAll(ctx context.Context) (model.CategoryList, error)
	Create(ctx context.Context, m *model.Category) (model.ID, error)
	Update(ctx context.Context, m *model.Category) error
	Delete(ctx context.Context, id model.ID) error
}

// DefaultCategoryService ...
type DefaultCategoryService struct {
	categoryRepo CategoryRepositoryInterface
}

// NewDefaultCategoryService create new default category service
func NewDefaultCategoryService(categoryRepo CategoryRepositoryInterface) *DefaultCategoryService {
	return &DefaultCategoryService{
		categoryRepo: categoryRepo,
	}
}

// ListCategories ...
func (s *DefaultCategoryService) ListCategories(ctx context.Context) (model.CategoryList, error) {
	return s.categoryRepo.FindAll(ctx)
}

// Create ...
func (s *DefaultCategoryService) Create(ctx context.Context, d *dto.Category) (model.ID, error) {

	m, err := s.makeModelFromCategoryDTO(d)
	if err != nil {
		return model.NilID, err
	}

	return s.categoryRepo.Create(ctx, m)
}

// Update ...
func (s *DefaultCategoryService) Update(ctx context.Context, d *dto.Category) error {

	m, err := s.makeModelFromCategoryDTO(d)
	if err != nil {
		return err
	}

	return s.categoryRepo.Update(ctx, m)
}

// Delete ...
func (s *DefaultCategoryService) Delete(ctx context.Context, id model.ID) error {
	return s.categoryRepo.Delete(ctx, id)
}

// Find ...
func (s *DefaultCategoryService) Find(ctx context.Context, id model.ID) (*model.Category, error) {
	return s.categoryRepo.Find(ctx, id)
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
