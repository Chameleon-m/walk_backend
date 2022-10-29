package dto

import (
	"github.com/go-playground/validator/v10"
)

func NewPlaceDTO() *Place {
	return &Place{}
}

type Place struct {
	ID          string   `json:"id" binding:"-"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Category    string   `json:"category" binding:"required"`
	Tags        []string `json:"tags"`
}

func ValidatePlaceDTO(sl validator.StructLevel) {

	place, ok := sl.Current().Interface().(Place)
	if !ok {
		return
	}

	if len(place.Name) < 5 {
		// place.Name -> sl.Current().Interface()
		sl.ReportError(sl.Current().Interface(), "name", "Name", "tag", "param")
	}
}

// https://github.com/go-playground/validator/blob/master/_examples/simple/main.go#L56
// https://www.golangprograms.com/go-struct-and-field-validation-examples.html
// https://blog.depa.do/post/gin-validation-errors-handling
