package repository

import (
	"encoding/json"
	"time"

	"walk_backend/internal/app/model"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

// PlaceCacheRedisRepository place mongodb repo
type PlaceCacheRedisRepository struct {
	client *redis.Client
}

// NewPlaceCacheRedisRepository create new redis place cache repository
func NewPlaceCacheRedisRepository(client *redis.Client) *PlaceCacheRedisRepository {
	return &PlaceCacheRedisRepository{
		client: client,
	}
}

// Get cache places
func (r *PlaceCacheRedisRepository) Get(ctx context.Context, key string) (model.PlaceList, error) {

	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	places := make(model.PlaceList, 0)
	if err = json.Unmarshal([]byte(result), &places); err != nil {
		return nil, err
	}
	return places, nil
}

// Set cache places
func (r *PlaceCacheRedisRepository) Set(ctx context.Context, key string, places model.PlaceList, expiration time.Duration) error {

	data, err := json.Marshal(places)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, string(data), expiration).Err()
}

// Del Delete cache places
func (r *PlaceCacheRedisRepository) Del(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}
