package service

import (
	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
)

// AuthServiceInteface ...
type AuthServiceInteface interface {
	Registration(dto *dto.AuthLogin) (*model.User, error)
	Login(dto *dto.AuthLogin) (*model.User, error)
	GenerateToken() (string, error)
}
