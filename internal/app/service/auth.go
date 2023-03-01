package service

import (
	"errors"

	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"

	"github.com/gofrs/uuid"
)

var (
	// ErrInvalidUsernameOrPassword ...
	ErrInvalidUsernameOrPassword = errors.New("invalid username or password")
)

// UserRepositoryInterface ...
type UserRepositoryInterface interface {
	Create(m *model.User) (model.ID, error)
	FindByUsername(username string) (*model.User, error)
}

// DefaultAuthService ...
type DefaultAuthService struct {
	userRepo UserRepositoryInterface
}

// NewDefaultAuthService create new default auth service
func NewDefaultAuthService(userRepo UserRepositoryInterface) *DefaultAuthService {
	return &DefaultAuthService{
		userRepo: userRepo,
	}
}

// Registration ...
func (s *DefaultAuthService) Registration(dto *dto.AuthLogin) (*model.User, error) {

	user, err := s.userRepo.FindByUsername(dto.Username)
	if err != nil && !errors.Is(err, model.ErrModelNotFound) {
		return nil, err
	} else if user != nil {
		return nil, ErrInvalidUsernameOrPassword
	}

	m, err := model.NewUserModel(dto.Username, dto.Password)
	if err != nil {
		return nil, err
	}

	if _, err := s.userRepo.Create(m); err != nil {
		return nil, err
	}

	return m, nil
}

// Login ...
func (s *DefaultAuthService) Login(dto *dto.AuthLogin) (*model.User, error) {

	user, err := s.userRepo.FindByUsername(dto.Username)
	if err != nil {
		if errors.Is(err, model.ErrModelNotFound) {
			return nil, ErrInvalidUsernameOrPassword
		}
		return nil, err
	}

	err = user.CheckPassword(dto.Password)
	if err != nil {
		if errors.Is(err, model.ErrPassMismatched) {
			return nil, ErrInvalidUsernameOrPassword
		}
		return nil, err
	}

	return user, nil
}

// GenerateToken ...
func (s *DefaultAuthService) GenerateToken() (string, error) {
	token, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return token.String(), nil
}
