package dto

func NewCategoryDTO() *Category {
	return &Category{}
}

type Category struct {
	ID    string `json:"id" binding:"-"`
	Name  string `json:"name" binding:"required"`
	Order int8   `json:"order" binding:"required"`
}
