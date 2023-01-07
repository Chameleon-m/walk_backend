package model

import (
	"time"
)

// NewPlaceModel create new place model
func NewPlaceModel(id ID, name string, nameSlug string, description string, category ID, tags []string) (*Place, error) {
	place := &Place{
		ID:          id,
		Name:        name,
		NameSlug:    nameSlug,
		Description: description,
		Category:    category,
		Tags:        tags,
	}
	if err := place.Validate(); err != nil {
		return nil, err
	}
	return place, nil
}

// Place ...
//
// swagger:parameters places newPlace
type Place struct {
	// swagger:ignore
	ID   ID     `bson:"_id"`
	Name string `bson:"name"`
	// swagger:ignore
	NameSlug    string   `bson:"nameSlug"`
	Description string   `bson:"description"`
	Category    ID       `bson:"category"`
	Tags        []string `bson:"tags"`

	// swagger:ignore
	CreatedAt time.Time `bson:"createdAt"`
	// swagger:ignore
	UpdatedAt time.Time `bson:"updatedAt,omitempty"`
	// swagger:ignore
	DeletedAt time.Time `bson:"deletedAt,omitempty"`
}

// PlaceList ...
type PlaceList []*Place

// Validate calidate place model
func (m *Place) Validate() error {

	if m.Name == "" || m.NameSlug == "" {
		return ErrInvalidModel
	}
	return nil
}
