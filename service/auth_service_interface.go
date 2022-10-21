package service

import (
	"walk_backend/dto"
	"walk_backend/model"
)

type AuthServiceInteface interface {
	Registration(dto *dto.AuthLogin) (*model.User, error)
	Login(dto *dto.AuthLogin) (*model.User, error)
	GenerateToken() (string, error)
}
