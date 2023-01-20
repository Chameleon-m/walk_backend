package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Auth middleware
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		session := sessions.Default(c)
		sessionToken := session.Get("token")
		if sessionToken == nil {
			c.AbortWithStatus(http.StatusForbidden)
		}
		c.Next()
	}
}
