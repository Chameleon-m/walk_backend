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
	ctx        context.Context
}

func NewUserMongoRepository(ctx context.Context, collection *mongo.Collection) *UserMongoRepository {
	return &UserMongoRepository{
		collection: collection,
		ctx:        ctx,
	}
}

func (r *UserMongoRepository) Create(m *model.User) (model.ID, error) {
	if m.ID.IsNil() {
		id, err := model.NewID()
		if err != nil {
			return model.NilID, err
		}
		m.ID = id
	}

	m.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(r.ctx, m)

	return m.ID, err
}

// Find user bu username
func (r *UserMongoRepository) FindByUsername(username string) (*model.User, error) {

	cur := r.collection.FindOne(r.ctx, bson.M{
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
