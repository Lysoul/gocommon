// Package gin is a helper package to get a gin compatible middleware.
package ginserver

import (
	"context"

	"github.com/gin-gonic/gin"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
)

func MetricsMiddleware(
	durationBuckets []float64,
	sizeBuckets []float64,
) gin.HandlerFunc {
	// Create our middleware.
	mdlw := middleware.New(middleware.Config{
		DisableMeasureInflight: true,
		Recorder: metrics.NewRecorder(metrics.Config{
			StatusCodeLabel: "http_status_code",
			HandlerIDLabel:  "http_route",
			MethodLabel:     "http_method",
			DurationBuckets: durationBuckets,
			// this is useless
			SizeBuckets: sizeBuckets,
		}),
	})
	return GinMetricsHandler("", mdlw)
}

// Handler returns a Gin measuring middleware.
func GinMetricsHandler(handlerID string, m middleware.Middleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		r := &reporter{c: c}
		m.Measure(handlerID, r, func() {
			c.Next()
		})
	}
}

type reporter struct {
	c *gin.Context
}

func (r *reporter) Method() string { return r.c.Request.Method }

func (r *reporter) Context() context.Context { return r.c.Request.Context() }

func (r *reporter) URLPath() string { return r.c.FullPath() }

func (r *reporter) StatusCode() int { return r.c.Writer.Status() }

func (r *reporter) BytesWritten() int64 { return int64(r.c.Writer.Size()) }
