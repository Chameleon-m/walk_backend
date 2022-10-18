package presenter

import (
	"walk_backend/model"
)

// Place list data
type Place struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    Category `json:"category"`
	Tags        []string `json:"tags"`
}

func NewPlacePresenter() *Place {
	return &Place{}
}

func (p Place) Make(m *model.Place, c *model.Category) *Place {
	p.ID = m.ID.String()
	p.Name = m.Name
	p.Description = m.Description
	p.Category = *p.Category.Make(c)
	p.Tags = m.Tags
	return &p
}

func (p *Place) MakeList(mList model.PlaceList, cList model.CategoryList) []*Place {

	list := make([]*Place, len(mList))
	for i := 0; i < len(mList); i++ {
		list[i] = p.Make(mList[i], cList.FindByID(mList[i].Category))
	}

	return list
}
