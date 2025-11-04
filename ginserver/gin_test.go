package ginserver_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lysoul/gocommon/ginserver"
	"github.com/Lysoul/gocommon/monitoring"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
)

func TestTrace(t *testing.T) {
	t.Setenv("LOG_MODE", "dev")
	t.Setenv("LOG_LEVEL", "DEBUG")
	t.Setenv("LOG_ENCODING", "console")

	zap.L().Info("test")

	resource, _ := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("foo"),
			attribute.String("library.language", "go"),
		),
	)

	cleaup := monitoring.InitTracer(&monitoring.TraceConfig{
		ServiceName:  "foobar",
		CollectorURL: "http://localhost:4317",
	}, resource)
	defer cleaup(context.Background())

	zapLogger, err := zap.NewDevelopment()
	require.NoError(t, err)
	otelZap := otelzap.New(zapLogger)

	r, _ := ginserver.InitGin(ginserver.Config{
		Mode:               "dev",
		Prefix:             "",
		CacheContolDefault: "",
	}, otelZap)

	// Example ping request.
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
	})

	reqID := "1234-5678-9012"
	req := httptest.NewRequest(http.MethodGet, "/ping?word=abc", nil)
	req.Header.Add("X-Request-ID", reqID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, reqID, res.Header.Get("X-Request-ID"))

	t.Run("test", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ping?word=abc", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		res := w.Result()
		defer res.Body.Close()

		require.NotEmpty(t, res.Header.Get("X-Request-ID"))
	})
}
