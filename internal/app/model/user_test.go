package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserModel(t *testing.T) {
	u, err := NewUserModel("Wozniak", "password")
	assert.Nil(t, err)
	assert.Equal(t, u.Username, "Wozniak")
	assert.NotNil(t, u.ID)
	assert.NotEqual(t, u.Password, "password")
}

func TestValidatePassword(t *testing.T) {
	u, _ := NewUserModel("Wozniak", "password")
	err := u.CheckPassword("password")
	assert.Nil(t, err)
	err = u.CheckPassword("wrong_password")
	assert.NotNil(t, err)
}

func TestUserValidate(t *testing.T) {
	type test struct {
		username string
		password string
		want     error
	}

	tests := []test{
		{
			username: "Wozniak",
			password: "password",
			want:     nil,
		},
		{
			username: "",
			password: "password",
			want:     ErrInvalidModel,
		},
		{
			username: "Wozniak",
			password: "",
			want:     ErrInvalidModel,
		},
		{
			username: "",
			password: "",
			want:     ErrInvalidModel,
		},
	}

	for _, tc := range tests {

		_, err := NewUserModel(tc.username, tc.password)
		assert.Equal(t, err, tc.want)
	}

}
