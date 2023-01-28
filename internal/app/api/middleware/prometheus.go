package middleware

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// Prometheus middleware
func Prometheus() gin.HandlerFunc {

	var totalRequests *prometheus.CounterVec
	var httpDuration *prometheus.HistogramVec
	var httpCodeCounter *prometheus.CounterVec

	g, gCtx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		totalRequests = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Number of incoming requests",
			},
			[]string{"path"},
		)

		return prometheus.Register(totalRequests)
	})

	g.Go(func() error {
		httpDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "http_response_time_seconds",
				Help: "Duration of HTTP requests",
			},
			[]string{"path", "method"},
		)

		return prometheus.Register(httpDuration)
	})

	g.Go(func() error {
		httpCodeCounter = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_response_status_total",
				Help: "Total number of reponses by HTTP status code.",
			},
			[]string{"code", "path", "method"},
		)

		return prometheus.Register(httpCodeCounter)
	})

	go func() {
		<-gCtx.Done()
		totalRequests, httpDuration, httpCodeCounter = nil, nil, nil
	}()

	if err := g.Wait(); err != nil {
		log.Error().Err(err).Send()
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(c.Request.URL.Path, c.Request.Method))
		totalRequests.WithLabelValues(c.Request.URL.Path).Inc()
		c.Next()
		timer.ObserveDuration()
		httpCodeCounter.WithLabelValues(strconv.Itoa(c.Writer.Status()), c.Request.URL.Path, c.Request.Method).Inc()
	}
}
