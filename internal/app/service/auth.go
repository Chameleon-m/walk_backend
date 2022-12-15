package service

import (
	"errors"

	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/app/repository"

	"github.com/gofrs/uuid"
)

var (
	ErrInvalidUsernameOrPassword = errors.New("Invalid username or password")
)

type DefaultAuthService struct {
	userRepo repository.UserRepositoryInterface
}

var _ AuthServiceInteface = (*DefaultAuthService)(nil)

func NewDefaultAuthService(userRepo repository.UserRepositoryInterface) *DefaultAuthService {
	return &DefaultAuthService{
		userRepo: userRepo,
	}
}

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

func (s *DefaultAuthService) GenerateToken() (string, error) {
	token, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return token.String(), nil
}
