package repository

import (
	"walk_backend/internal/app/model"
)

type UserRepositoryInterface interface {
	Create(m *model.User) (model.ID, error)
	FindByUsername(username string) (*model.User, error)
}