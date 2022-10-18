package presenter

import (
	"walk_backend/model"
)

type Category struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int8   `json:"order"`
}

func NewCategoryPresenter() *Category {
	return &Category{}
}

func (p Category) Make(m *model.Category) *Category {
	p.ID = m.ID.String()
	p.Name = m.Name
	p.Order = m.Order
	return &p
}

func (p *Category) MakeList(mList model.CategoryList) []*Category {

	list := make([]*Category, 0, len(mList))
	for _, m := range mList {
		list = append(list, p.Make(m))
	}

	return list
}
