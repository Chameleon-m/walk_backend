package repository

import (
	"walk_backend/internal/app/model"
)

type PlaceQueueRepositoryInterface interface {
	PublishReIndex(id model.ID) error
}
