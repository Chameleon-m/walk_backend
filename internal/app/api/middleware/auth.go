package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {

		session := sessions.Default(c)
		sessionToken := session.Get("token")
		tokenValue := c.GetHeader("Authorization")
		if sessionToken == nil {
			c.AbortWithStatus(http.StatusForbidden)
		} else if tokenValue != sessionToken {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		c.Next()
	}
}
