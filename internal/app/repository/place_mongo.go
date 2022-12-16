package repository

import (
	"errors"
	"time"

	"walk_backend/internal/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

// PlaceMongoRepository place mongodb repo
type PlaceMongoRepository struct {
	collection *mongo.Collection
	ctx        context.Context
}

var _ PlaceRepositoryInterface = (*PlaceMongoRepository)(nil)

func NewPlaceMongoRepository(ctx context.Context, collection *mongo.Collection) *PlaceMongoRepository {
	return &PlaceMongoRepository{
		collection: collection,
		ctx:        ctx,
	}
}

// Find place
func (r *PlaceMongoRepository) Find(id model.ID) (*model.Place, error) {

	cur := r.collection.FindOne(r.ctx, bson.M{
		"_id": id,
	})

	if cur.Err() != nil {
		if errors.Is(cur.Err(), mongo.ErrNoDocuments) {
			return nil, model.ErrModelNotFound
		}
		return nil, cur.Err()
	}

	var m model.Place
	if err := cur.Decode(&m); err != nil {
		return nil, err
	}

	return &m, nil
}

// FindAll places
func (r *PlaceMongoRepository) FindAll() (model.PlaceList, error) {

	cursor, err := r.collection.Find(r.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	mList := make(model.PlaceList, 0)
	for cursor.Next(r.ctx) {
		var place model.Place
		if err := cursor.Decode(&place); err != nil {
			return nil, err
		}
		mList = append(mList, &place)
	}

	return mList, nil
}

// Create
func (r *PlaceMongoRepository) Create(place *model.Place) (model.ID, error) {

	if place.ID.IsNil() {
		id, err := model.NewID()
		if err != nil {
			return model.NilID, err
		}
		place.ID = id
	}

	place.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(r.ctx, place)

	return place.ID, err
}

// Update
func (r *PlaceMongoRepository) Update(place *model.Place) error {

	place.UpdatedAt = time.Now()

	updateResult, err := r.collection.UpdateOne(r.ctx, bson.M{
		"_id": place.ID,
	}, bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: place.Name},
		{Key: "description", Value: place.Description},
		{Key: "category", Value: place.Category},
		{Key: "tags", Value: place.Tags},
		{Key: "updatedAt", Value: place.UpdatedAt},
	}}})

	if updateResult.MatchedCount == 0 {
		return model.ErrModelNotFound
	} else if updateResult.ModifiedCount == 0 {
		return model.ErrModelUpdate
	}

	return err
}

// Delete
func (r *PlaceMongoRepository) Delete(id model.ID) error {
	deleteResult, err := r.collection.DeleteOne(r.ctx, bson.M{
		"_id": id,
	})

	if deleteResult.DeletedCount == 0 {
		return model.ErrModelNotFound
	}

	return err
}

func (r *PlaceMongoRepository) Search(search string) (model.PlaceList, error) {

	sort := options.Find()
	sort.SetSort(bson.D{{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}}})
	cursor, err := r.collection.Find(r.ctx, bson.D{{Key: "$text", Value: bson.D{{Key: "$search", Value: search}}}}, sort)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	places := make(model.PlaceList, 0)
	for cursor.Next(r.ctx) {
		var place model.Place
		if err := cursor.Decode(&place); err != nil {
			return nil, err
		}
		places = append(places, &place)
	}

	return places, nil
}
