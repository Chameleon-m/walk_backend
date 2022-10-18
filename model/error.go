package model

import (
	"errors"
)

var (
	ErrInvalidString = errors.New("The provided string is not a valid ID")
	ErrModelNotFound = errors.New("Model not found")
	ErrModelUpdate   = errors.New("No upsert was done")
	ErrInvalidEntity = errors.New("Invalid entity")
)

// IsErrInvalidString check is a ErrInvalidString
func IsErrInvalidString(err error) bool {
	return errors.Is(err, ErrInvalidString)
}

// IsErrModelNotFound check is a ErrModelNotFound
func IsErrModelNotFound(err error) bool {
	return errors.Is(err, ErrModelNotFound)
}

// IsErrModelUpdate check is a ErrModelUpdate
func IsErrModelUpdate(err error) bool {
	return errors.Is(err, ErrModelUpdate)
}

// IsErrModelUpdate check is a ErrInvalidEntity
func IsErrInvalidEntity(err error) bool {
	return errors.Is(err, ErrInvalidEntity)
}