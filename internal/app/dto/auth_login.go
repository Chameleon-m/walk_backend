package dto

// NewAuthLoginDTO create new auth login DTO
func NewAuthLoginDTO() *AuthLogin {
	return &AuthLogin{}
}

// AuthLogin ...
type AuthLogin struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
