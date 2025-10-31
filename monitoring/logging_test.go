package monitoring_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestLoggingConfig(t *testing.T) {
	t.Setenv("LOG_MODE", "dev")
	t.Setenv("LOG_LEVEL", "DEBUG")

	level, err := zap.ParseAtomicLevel("DEBUG")
	if err != nil {
		t.Fatal(err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	zapconf := zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr", cwd + "/test.log"},
		ErrorOutputPaths: []string{"stderr", cwd + "/test.log"},
	}
	l, err := zapconf.Build()
	require.NoError(t, err)
	logger := otelzap.New(l)

	tracer := otel.Tracer("app_or_package_name")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, "root")
	defer span.End()

	logger.Ctx(ctx).Error("test", zap.Error(errors.New("test error")))
	logger.Ctx(ctx).Warn("warn")
	logger.Ctx(ctx).Info("info")
	span2 := trace.SpanFromContext(ctx)
	// Save the active span in the context.
	ctx = trace.ContextWithSpan(ctx, span2)
	logger.Ctx(ctx).Debug("debug")
	span2.End()
	span.RecordError(errors.New("test error"))
}
