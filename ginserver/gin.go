package ginserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const (
	xRequestIDKey = "X-Request-ID"
)

// we could have a stuct with these things attached to it
// but whats the point...
//
//nolint:gochecknoglobals // we want to have a global config.
var config Config

type Config struct {
	Mode    string   `envconfig:"HTTP_MODE" default:"debug"`
	LogSkip []string `envconfig:"HTTP_LOG_SKIP" default:"/health,/metrics"`
	Port    int      `envconfig:"HTTP_PORT" default:"3000"`
	Prefix  string   `envconfig:"HTTP_PATH_PREFIX" default:""`

	EnableCORS          bool          `envconfig:"HTTP_ENABLE_CORS" default:"false"`
	AllowedOrigins      []string      `envconfig:"HTTP_ALLOWED_ORIGINS" default:"*"`
	AllowedHeaders      []string      `envconfig:"HTTP_ALLOWED_HEADERS" default:"*"`
	CORSMaxAge          time.Duration `envconfig:"HTTP_CORS_MAX_AGE" default:"24h"`
	DefaultCacheControl string        `envconfig:"HTTP_DEFAULT_CACHE_CONTROL" default:"no-store"`

	// generate etag for all responses
	EnableEtag bool `envconfig:"HTTP_ENABLE_ETAG" default:"false"`

	// default value to put in cache control, leave empty to disable
	// `private, no-store, no-cache` are good options
	CacheContolDefault string `envconfig:"HTTP_CACHE_CONTROL_DEFAULT"`

	ServiceName string `envconfig:"SERVICE_NAME" default:"service"`

	MetricsDurationBuckets []float64 `envconfig:"HTTP_METRICS_DURATION_BUCKETS" default:"0.2,0.5,1,3"`
	MetricsSizeBuckets     []float64 `envconfig:"HTTP_METRICS_SIZE_BUCKETS" default:"100,1000,10000,100000"`
}

// returns router and function to run server.
func InitGin(_config Config, logger *otelzap.Logger) (*gin.Engine, func() (*http.Server, func(context.Context))) {
	config = _config

	if config.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	router := gin.New()
	router.ContextWithFallback = true

	if config.EnableCORS {
		logger.Info("CORS enabled")
		router.Use(CORSMiddleware(config))
	} else {
		logger.Info("CORS disabled")
	}

	router.Use(RequestID())
	router.Use(otelgin.Middleware(config.ServiceName))
	router.Use(
		GinZapSetup(logger, config),
		// Logs all panic to error log
		//   - stack means whether output the stack info.
		ginzap.RecoveryWithZap(logger.Logger, true),
	)

	router.Use(MetricsMiddleware(
		config.MetricsDurationBuckets,
		config.MetricsSizeBuckets,
	))

	if config.CacheContolDefault != "" {
		router.Use(CacheControlMiddleware(config.CacheContolDefault))
	}

	return router, func() (*http.Server, func(context.Context)) {
		return Run(router)
	}
}

// to do return servver AND shutdown.
func Run(router *gin.Engine) (*http.Server, func(context.Context)) {
	var handler http.Handler
	if config.EnableEtag {
		handler = EtagHandler(router)
	} else {
		handler = router
	}
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", config.Port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && errors.Is(err, http.ErrServerClosed) {
			zap.L().Info("ListenAndServe", zap.Error(err))
		}
	}()

	return server, func(ctx context.Context) {
		err := server.Shutdown(ctx)
		if err != nil {
			zap.L().Error("Server forced to shutdown:", zap.Error(err))
		} else {
			zap.L().Info("HTTP server has shut down")
		}
	}
}

func GinZapSetup(
	logger *otelzap.Logger,
	config Config,
) gin.HandlerFunc {
	return ginzap.GinzapWithConfig(
		logger.Logger,
		&ginzap.Config{
			TimeFormat: time.RFC3339,
			// UTC:        true,
			SkipPaths: config.LogSkip,
			Context: ginzap.Fn(func(c *gin.Context) []zapcore.Field {
				fields := []zapcore.Field{}
				// log request ID
				if requestID := c.Request.Header.Get("X-Request-Id"); requestID != "" {
					fields = append(fields, zap.String("request_id", requestID))
				}
				// log trace and span ID
				if trace.SpanFromContext(c.Request.Context()).SpanContext().IsValid() {
					fields = append(fields,
						zap.String("trace_id",
							trace.SpanFromContext(c.Request.Context()).SpanContext().TraceID().String()))
					fields = append(fields,
						zap.String("span_id",
							trace.SpanFromContext(c.Request.Context()).SpanContext().SpanID().String()))
				}
				return fields
			}),
		},
	)
}
