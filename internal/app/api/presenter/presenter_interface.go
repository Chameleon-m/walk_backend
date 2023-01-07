package presenter

import (
	"walk_backend/internal/app/model"
)

// PlacePresenterInteface ...
type PlacePresenterInteface interface {
	Make(m *model.Place, c *model.Category) *Place
	MakeList(mList model.PlaceList, cList model.CategoryList) []*Place
}

// CategoryPresenterInteface ...
type CategoryPresenterInteface interface {
	Make(m *model.Category) *Category
	MakeList(mList model.CategoryList) []*Category
}

// TokenPresenterInteface ...
type TokenPresenterInteface interface {
	Make(token string) *Token
}
