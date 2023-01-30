package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Number of incoming requests",
	},
	[]string{"path"},
)

var httpDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_response_time_seconds",
		Help: "Duration of HTTP requests",
	},
	[]string{"path", "method"},
)

var httpCodeCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_response_status_total",
		Help: "Total number of reponses by HTTP status code.",
	},
	[]string{"code", "path", "method"},
)

func init() {
	prometheus.MustRegister(totalRequests, httpDuration, httpCodeCounter)
}

// Prometheus middleware
func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(c.Request.URL.Path, c.Request.Method))
		totalRequests.WithLabelValues(c.Request.URL.Path).Inc()
		c.Next()
		timer.ObserveDuration()
		httpCodeCounter.WithLabelValues(strconv.Itoa(c.Writer.Status()), c.Request.URL.Path, c.Request.Method).Inc()
	}
}
