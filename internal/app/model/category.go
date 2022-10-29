package model

type Category struct {
	ID    ID     `bson:"_id"`
	Name  string `bson:"name"`
	Order int8   `bson:"order"`
}

type CategoryList []*Category

func NewCategoryModel(id ID, name string, order int8) (*Category, error) {
	m := &Category{
		ID:    id,
		Name:  name,
		Order: order,
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Category) Validate() error {

	if m.Name == "" || m.Order == 0 {
		return ErrInvalidModel
	}
	return nil
}

func (mL CategoryList) FindByID(id ID) *Category {
	for _, m := range mL {
		if m.ID == id {
			return m
		}
	}

	return nil
}
