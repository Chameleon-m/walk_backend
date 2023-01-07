package repository

import (
	"walk_backend/internal/app/model"
)

// PlaceQueueRepositoryInterface ...
type PlaceQueueRepositoryInterface interface {
	PublishReIndex(id model.ID) error
}
