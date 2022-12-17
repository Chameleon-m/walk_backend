package model

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// NewUserModel create new user model
func NewUserModel(username string, password string) (*User, error) {
	id, err := NewID()
	if err != nil {
		return nil, err
	}

	m := &User{
		ID:       id,
		Username: username,
		Password: password,
	}

	if err := m.Validate(); err != nil {
		return nil, err
	}

	pwd, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	m.Password = pwd

	return m, nil
}

// User ...
//
// swagger:parameters auth signIn
type User struct {
	// swagger:ignore
	ID       ID     `bson:"_id"`
	Username string `bson:"username"`
	Password string `bson:"password"`
	// swagger:ignore
	CreatedAt time.Time `bson:"createdAt"`
}

// Validate calidate user model
func (m *User) Validate() error {

	if m.Username == "" || m.Password == "" {
		return ErrInvalidModel
	}
	return nil
}

// CheckPassword check user password
func (m *User) CheckPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(password))
	if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrPassMismatched
	}
	return err
}

func hashPassword(raw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}
