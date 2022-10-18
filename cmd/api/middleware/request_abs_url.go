package middleware

import (
	"github.com/gin-gonic/gin"
)

func RequestAbsUrl() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.IsAbs() {
			return
		}

		url := c.Request.URL

		if url.Scheme == "" {
			xForwardedProto := c.Request.Header.Get("X-Forwarded-Proto")
			if xForwardedProto != "" {
				url.Scheme = xForwardedProto
			} else {
				if c.IsWebsocket() {
					url.Scheme = "ws"
					if c.Request.TLS != nil {
						url.Scheme = "wss"
					}

				} else {
					url.Scheme = "http"
					if c.Request.TLS != nil {
						url.Scheme = "https"
					}
				}
			}
		}

		if url.Host == "" {
			url.Host = c.Request.Host
		}
	}
}
