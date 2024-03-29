package repository

import (
	"errors"
	"time"

	"walk_backend/internal/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

// UserMongoRepository category mongodb repo
type UserMongoRepository struct {
	collection *mongo.Collection
}

// NewUserMongoRepository create new user mongo repository
func NewUserMongoRepository(collection *mongo.Collection) *UserMongoRepository {
	return &UserMongoRepository{
		collection: collection,
	}
}

// Create ...
func (r *UserMongoRepository) Create(ctx context.Context, m *model.User) (model.ID, error) {
	if m.ID.IsNil() {
		id, err := model.NewID()
		if err != nil {
			return model.NilID, err
		}
		m.ID = id
	}

	m.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(ctx, m)

	return m.ID, err
}

// FindByUsername user bu username
func (r *UserMongoRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {

	cur := r.collection.FindOne(ctx, bson.M{
		"username": username,
	})

	if cur.Err() != nil {
		if errors.Is(cur.Err(), mongo.ErrNoDocuments) {
			return nil, model.ErrModelNotFound
		}
		return nil, cur.Err()
	}

	var m model.User
	if err := cur.Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}
