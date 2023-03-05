package middleware

import (
	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// RequestLogger middleware
func RequestLogger(log *zerolog.Logger, skipPath []string) gin.HandlerFunc {
	return logger.SetLogger(
		logger.WithSkipPath(skipPath),
		// logger.WithSkipPathRegexp(nil),
		logger.WithUTC(true),
		// logger.WithWriter(os.Stdout),
		// logger.WithDefaultLevel(zerolog.InfoLevel),
		// logger.WithClientErrorLevel(zerolog.WarnLevel),
		// logger.WithServerErrorLevel(zerolog.ErrorLevel),
		logger.WithLogger(func(c *gin.Context, l zerolog.Logger) zerolog.Logger {
			return log.With().Str("id", c.GetHeader("X-Request-ID")).Logger()
		}),
	)
}
