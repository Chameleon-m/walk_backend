package handlers

import (
	"errors"
	"net/http"

	"walk_backend/internal/app/api/presenter"
	"walk_backend/internal/app/dto"
	"walk_backend/internal/app/service"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/context"
)

// AuthHandler auth handler
type AuthHandler struct {
	service   service.AuthServiceInteface
	ctx       context.Context
	presenter presenter.TokenPresenterInteface
}

// NewAuthHandler create new auth handler
func NewAuthHandler(ctx context.Context, service service.AuthServiceInteface, presenter presenter.TokenPresenterInteface) *AuthHandler {
	return &AuthHandler{
		service:   service,
		ctx:       ctx,
		presenter: presenter,
	}
}

// SignUpHandler registration
//
// swagger:operation POST /auth/registration auth signUp
// Registration with username and password
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	  description: Successful operation
//	'400':
//	  description: Invalid input
//	'500':
//	  description: Invalid credentials
func (handler *AuthHandler) SignUpHandler(c *gin.Context) {

	dto := dto.NewAuthLoginDTO()
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := handler.service.Registration(dto)
	if err != nil {
		c.Error(err)
		if errors.Is(err, service.ErrInvalidUsernameOrPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, "Auth service registration error")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User signed up"})
}

// SignInHandler login
//
// swagger:operation POST /auth/login auth signIn
// Login with username and password
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	  description: Successful operation
//	'401':
//	  description: Invalid credentials
//	'500':
//	  description: Status Internal Server
func (handler *AuthHandler) SignInHandler(c *gin.Context) {

	dto := dto.NewAuthLoginDTO()
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := handler.service.Login(dto)
	if err != nil {
		c.Error(err)
		if errors.Is(err, service.ErrInvalidUsernameOrPassword) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, "Auth service login error")
		return
	}

	sessionTokenNew, err := handler.service.GenerateToken()
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generate token"})
		return
	}
	session := sessions.Default(c)
	session.Set("username", user.Username)
	session.Set("token", sessionTokenNew)
	session.Save()

	data := handler.presenter.Make(sessionTokenNew)
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// RefreshHandler refresh token
//
// swagger:operation POST /auth/refresh-tokens auth refresh
// Refresh token
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	  description: Successful operation
//	'401':
//	  description: Invalid credentials
func (handler *AuthHandler) RefreshHandler(c *gin.Context) {

	session := sessions.Default(c)
	sessionToken := session.Get("token")
	if sessionToken == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session cookie"})
		return
	}

	sessionTokenNew, err := handler.service.GenerateToken()
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generate token"})
		return
	}
	session.Set("token", sessionTokenNew)
	session.Save()

	c.JSON(http.StatusOK, gin.H{"message": "New session issued"})
}

// SignOutHandler logout
//
// swagger:operation POST /auth/logout auth signOut
// Signing out
// ---
// responses:
//
//	'200':
//	  description: Successful operation
func (handler *AuthHandler) SignOutHandler(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.JSON(http.StatusOK, gin.H{"message": "Signed out..."})
}

// MakeHandlers make handlers
func (handler *AuthHandler) MakeHandlers(router *gin.RouterGroup) {

	router.POST("/auth/registration", handler.SignUpHandler)
	router.POST("/auth/login", handler.SignInHandler)
	router.POST("/auth/refresh-tokens", handler.RefreshHandler)
	router.POST("/auth/logout", handler.SignOutHandler)
}
