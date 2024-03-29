package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	authMock "walk_backend/internal/app/api/handlers/auth/mock"
	"walk_backend/internal/app/api/presenter"
	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/model"
	"walk_backend/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Registration(t *testing.T) {

	credentialsCase := []dto.AuthLogin{
		{Username: "test", Password: "test"},
		{Username: "", Password: "test"},
		{Username: "test", Password: ""},
		{Username: "test", Password: "test"},
		{Username: "test", Password: "test"},
	}

	url := "/api/v1/auth/registration"

	controller := gomock.NewController(t)
	defer controller.Finish()

	router := gin.Default()
	apiV1 := router.Group("/api/v1")

	mockAuthService := authMock.NewMockServiceInterface(controller)

	mh := NewHandler(context.Background(), apiV1, mockAuthService, presenter.NewTokenPresenter())
	mh.MakeRoutes()

	t.Run("Ok", func(t *testing.T) {

		user, err := model.NewUserModel(credentialsCase[0].Username, credentialsCase[0].Password)
		assert.Nil(t, err)

		mockAuthService.
			EXPECT().
			Registration(context.Background(), &credentialsCase[0]).
			Return(user, nil).
			Times(1)

		jsonCredentials, _ := json.Marshal(credentialsCase[0])

		// Registration
		request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonCredentials))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("Empty_username_dto_validate", func(t *testing.T) {

		jsonCredentials, _ := json.Marshal(credentialsCase[1])

		request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonCredentials))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("Empty_password_dto_validate", func(t *testing.T) {

		jsonCredentials, _ := json.Marshal(credentialsCase[2])

		request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonCredentials))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("Fail_user_exist", func(t *testing.T) {

		mockAuthService.
			EXPECT().
			Registration(context.Background(), &credentialsCase[3]).
			Return(nil, service.ErrInvalidUsernameOrPassword).
			Times(1)

		jsonCredentials, _ := json.Marshal(credentialsCase[3])

		request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonCredentials))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("Fail_user_model_validate", func(t *testing.T) {

		mockAuthService.
			EXPECT().
			Registration(context.Background(), &credentialsCase[4]).
			Return(nil, model.ErrInvalidModel).
			Times(1)

		jsonCredentials, _ := json.Marshal(credentialsCase[4])

		request, _ := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonCredentials))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}
