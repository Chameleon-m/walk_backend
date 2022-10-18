package repository

import (
	"errors"
	"fmt"

	"walk_backend/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

// CategoryMongoRepository category mongodb repo
type CategoryMongoRepository struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewCategoryMongoRepository(ctx context.Context, collection *mongo.Collection) *CategoryMongoRepository {
	return &CategoryMongoRepository{
		collection: collection,
		ctx:        ctx,
	}
}

// Find category
func (r *CategoryMongoRepository) Find(id model.ID) (*model.Category, error) {

	cur := r.collection.FindOne(r.ctx, bson.M{
		"_id": id,
	})

	if cur.Err() != nil {
		if errors.Is(cur.Err(), mongo.ErrNoDocuments) {
			return nil, model.ErrModelNotFound
		}
		return nil, cur.Err()
	}

	var m model.Category
	if err := cur.Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

// FindAll categories
func (r *CategoryMongoRepository) FindAll() (model.CategoryList, error) {

	cursor, err := r.collection.Find(r.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	mList := make(model.CategoryList, 0)
	for cursor.Next(r.ctx) {
		var m model.Category
		if err := cursor.Decode(&m); err != nil {
			return nil, err
		}
		mList = append(mList, &m)
	}

	return mList, nil
}

// Create
func (r *CategoryMongoRepository) Create(m *model.Category) (model.ID, error) {

	if m.ID.IsNil() {
		id, err := model.NewID()
		if err != nil {
			return model.NilID, err
		}
		m.ID = id
	}

	_, err := r.collection.InsertOne(r.ctx, m)

	return m.ID, err
}

// Update
func (r *CategoryMongoRepository) Update(m *model.Category) error {

	fmt.Println(m)
	updateResult, err := r.collection.UpdateOne(r.ctx, bson.M{
		"_id": m.ID,
	}, bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: m.Name},
		{Key: "order", Value: m.Order},
	}}})

	if updateResult.MatchedCount == 0 {
		return model.ErrModelNotFound
	} else if updateResult.ModifiedCount == 0 {
		return model.ErrModelUpdate
	}

	return err
}

// Delete
func (r *CategoryMongoRepository) Delete(id model.ID) error {
	deleteResult, err := r.collection.DeleteOne(r.ctx, bson.M{
		"_id": id,
	})

	if deleteResult.DeletedCount == 0 {
		return model.ErrModelNotFound
	}

	return err
}