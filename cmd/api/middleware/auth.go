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
		if sessionToken == nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		tokenValue := c.GetHeader("Authorization")
		if tokenValue != sessionToken {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}
