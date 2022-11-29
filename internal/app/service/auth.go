package service

import (
	"errors"

	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/app/repository"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/bcrypt"
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

// TODO
func (s *DefaultAuthService) Registration(dto *dto.AuthLogin) (*model.User, error) {

	user, err := s.userRepo.FindByUsername(dto.Username)
	// TODO
	if err != nil && !errors.Is(err, model.ErrModelNotFound) {
		return nil, err
	} else if user != nil {
		return nil, ErrInvalidUsernameOrPassword
	}

	hashPassword, err := hashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	id, err := model.NewID()
	if err != nil {
		return nil, err
	}

	m, err := model.NewUserModel(id, dto.Username, hashPassword)
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

	hashPassword, err := hashPassword(dto.Password)
	if err != nil {
		return nil, err
	} else if checkPasswordHash(user.Password, hashPassword) {
		return nil, ErrInvalidUsernameOrPassword
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

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
