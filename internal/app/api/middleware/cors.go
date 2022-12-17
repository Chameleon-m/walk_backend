package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Cors middleware
func Cors(siteSchema string, siteHost string, sitePort string) gin.HandlerFunc {
	AllowOrigins := []string{siteSchema + "://" + siteHost + ":" + sitePort}
	return cors.New(cors.Config{
		AllowOrigins:     AllowOrigins,
		AllowMethods:     []string{"OPTIONS", "GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
