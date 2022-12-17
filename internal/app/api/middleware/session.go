package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Session middleware
func Session(name string, store sessions.Store) gin.HandlerFunc {
	return sessions.Sessions(name, store)
}
