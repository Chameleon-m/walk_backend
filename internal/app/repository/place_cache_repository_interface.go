package repository

import (
	"time"

	"walk_backend/internal/app/model"
)

// PlaceCacheRepositoryInterface ...
type PlaceCacheRepositoryInterface interface {
	Get(key string) (model.PlaceList, error)
	Set(key string, value model.PlaceList, expiration time.Duration) error
	Del(keys ...string) error
}
