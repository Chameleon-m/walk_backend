package repository

import (
	"walk_backend/model"
)

type PlaceQueueRepositoryInterface interface {
	PublishReIndex(id model.ID) error
}
