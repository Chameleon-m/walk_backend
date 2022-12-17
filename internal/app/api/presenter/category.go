package presenter

import (
	"walk_backend/internal/app/model"
)

// Category ...
type Category struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int8   `json:"order"`
}

// NewCategoryPresenter creaete new category presenter
func NewCategoryPresenter() *Category {
	return &Category{}
}

// Make make category presenter
func (p Category) Make(m *model.Category) *Category {
	p.ID = m.ID.String()
	p.Name = m.Name
	p.Order = m.Order
	return &p
}

// MakeList make category presenter list
func (p *Category) MakeList(mList model.CategoryList) []*Category {

	list := make([]*Category, 0, len(mList))
	for _, m := range mList {
		list = append(list, p.Make(m))
	}

	return list
}
