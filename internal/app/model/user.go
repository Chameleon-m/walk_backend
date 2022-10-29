package model

import (
	"time"
)

func NewUserModel(id ID, username string, password string) (*User, error) {
	m := &User{
		ID:       id,
		Username: username,
		Password: password,
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return m, nil
}

// swagger:parameters auth signIn
type User struct {
	// swagger:ignore
	ID       ID     `bson:"_id"`
	Username string `bson:"username"`
	Password string `bson:"password"`
	// swagger:ignore
	CreatedAt time.Time `bson:"createdAt"`
}

func (m *User) Validate() error {

	if m.Username == "" || m.Password == "" {
		return ErrInvalidModel
	}
	return nil
}
