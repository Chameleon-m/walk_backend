package service

import (
	"testing"

	"walk_backend/internal/app/dto"

	"walk_backend/internal/app/model"
	mockRepository "walk_backend/internal/app/repository/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	credentials dto.AuthLogin
	err         error
}

func TestAuthService_Registration(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockUserRepository := mockRepository.NewMockUserRepositoryInterface(controller)

	t.Run("ErrInvalidUsernameOrPassword", func(t *testing.T) {

		testCase := testCase{
			credentials: dto.AuthLogin{Username: "test", Password: "test"},
			err:         ErrInvalidUsernameOrPassword,
		}

		userModel, _ := model.NewUserModel(
			testCase.credentials.Username,
			testCase.credentials.Password,
		)

		mockUserRepository.
			EXPECT().
			FindByUsername(testCase.credentials.Username).
			Return(userModel, nil).
			Times(1)

		das := NewDefaultAuthService(mockUserRepository)
		_, err := das.Registration(&testCase.credentials)
		assert.ErrorIs(t, err, testCase.err)
	})
}

func TestAuthService_Login(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockUserRepository := mockRepository.NewMockUserRepositoryInterface(controller)

	t.Run("ErrInvalidUsernameOrPassword", func(t *testing.T) {

		testCases := []testCase{
			{
				credentials: dto.AuthLogin{Username: "test", Password: "test"},
				err:         ErrInvalidUsernameOrPassword,
			},
		}

		for _, testCase := range testCases {

			mockUserRepository.
				EXPECT().
				FindByUsername(testCase.credentials.Username).
				Return(nil, model.ErrModelNotFound).
				Times(1)

			das := NewDefaultAuthService(mockUserRepository)
			_, err := das.Login(&testCase.credentials)
			assert.ErrorIs(t, err, testCase.err)
		}
	})
}