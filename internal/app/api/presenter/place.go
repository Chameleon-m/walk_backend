package presenter

import (
	"walk_backend/internal/app/model"
)

// Place list data
type Place struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    Category `json:"category"`
	Tags        []string `json:"tags"`
}

// NewPlacePresenter create new place presenter
func NewPlacePresenter() *Place {
	return &Place{}
}

// Make make place presenter
func (p Place) Make(m *model.Place, c *model.Category) *Place {
	p.ID = m.ID.String()
	p.Name = m.Name
	p.Description = m.Description
	p.Category = *p.Category.Make(c)
	p.Tags = m.Tags
	return &p
}

// MakeList make list place prersenters
func (p *Place) MakeList(mList model.PlaceList, cList model.CategoryList) []*Place {

	list := make([]*Place, len(mList))
	for i := 0; i < len(mList); i++ {
		list[i] = p.Make(mList[i], cList.FindByID(mList[i].Category))
	}

	return list
}
