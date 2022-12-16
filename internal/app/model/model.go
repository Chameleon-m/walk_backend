package model

import (
	"github.com/gofrs/uuid"
)

// ID entity ID
type ID = uuid.UUID

// NilID is the zero value for ID.
var NilID = uuid.Nil

// Criteria TODO Interface
type Criteria any

// NewID create a new entity ID
func NewID() (ID, error) {
	return uuid.NewV7()
}

// StringToID convert a string to an entity ID
func StringToID(s string) (ID, error) {
	id, err := uuid.FromString(s)
	return ID(id), err
}
