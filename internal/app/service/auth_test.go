package service

import (
	"context"
	"testing"

	"walk_backend/internal/app/dto"

	"walk_backend/internal/app/model"
	mockAuth "walk_backend/internal/app/service/mock"

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

	mockUserRepository := mockAuth.NewMockUserRepositoryInterface(controller)

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
			FindByUsername(context.Background(), testCase.credentials.Username).
			Return(userModel, nil).
			Times(1)

		das := NewDefaultAuthService(mockUserRepository)
		_, err := das.Registration(context.Background(), &testCase.credentials)
		assert.ErrorIs(t, err, testCase.err)
	})
}

func TestAuthService_Login(t *testing.T) {

	controller := gomock.NewController(t)
	defer controller.Finish()

	mockUserRepository := mockAuth.NewMockUserRepositoryInterface(controller)

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
				FindByUsername(context.Background(), testCase.credentials.Username).
				Return(nil, model.ErrModelNotFound).
				Times(1)

			das := NewDefaultAuthService(mockUserRepository)
			_, err := das.Login(context.Background(), &testCase.credentials)
			assert.ErrorIs(t, err, testCase.err)
		}
	})
}
