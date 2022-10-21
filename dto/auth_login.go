package dto

func NewAuthLoginDTO() *AuthLogin {
	return &AuthLogin{}
}

type AuthLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
