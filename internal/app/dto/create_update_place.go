package dto

import (
	"github.com/go-playground/validator/v10"
)

// NewPlaceDTO create new place DTO
func NewPlaceDTO() *Place {
	return &Place{}
}

// Place ...
type Place struct {
	ID          string   `json:"id" binding:"-"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Category    string   `json:"category" binding:"required"`
	Tags        []string `json:"tags"`
}

// ValidatePlaceDTO validate place DTO
func ValidatePlaceDTO(sl validator.StructLevel) {

	place, ok := sl.Current().Interface().(Place)
	if !ok {
		return
	}

	if len(place.Name) < 5 {
		sl.ReportError(sl.Current().Interface(), "name", "Name", "tag", "param")
	}
}
