package monitoring

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

// returns graceful shutdown function (basically trace shutdown).
func InitTelemetry(config *TraceConfig) func(ctx context.Context) error {
	resource, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		panic(err)
	}
	InitMetrics(resource)
	return InitTracer(config, resource)
}

func ServeTelemetry(port int) {
	h := http.NewServeMux()

	// register a new handler for the /metrics endpoint
	h.Handle("/metrics", promhttp.Handler())
	h.Handle("/health", HealthHandler(HealthConfig{}))
	// start an http server using the mux server
	server := http.Server{
		Handler:           h,
		Addr:              ":" + fmt.Sprint(port),
		ReadHeaderTimeout: 1 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}

// direclty set
// OTEL_TRACES_SAMPLER
// OTEL_TRACES_SAMPLER_ARG.
type TraceConfig struct {
	Environment  string `envconfig:"ENVIRONMENT" default:"local"`
	ServiceName  string `envconfig:"SERVICE_NAME" default:"service"`
	CollectorURL string `envconfig:"TRACE_COLLECTOR_URL" default:"localhost:4317"`
	Insecure     bool   `envconfig:"TRACE_INSECURE" default:"false"`
}

func InitMetrics(resource *resource.Resource) {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal("failed to initialize prometheus exporter", zap.Error(err))
	}
	provider := metric.NewMeterProvider(metric.WithReader(exporter), metric.WithResource(resource))
	otel.SetMeterProvider(provider)
}

func InitTracer(config *TraceConfig, resource *resource.Resource) func(ctx context.Context) error {
	if config == nil {
		config = &TraceConfig{}
		envconfig.MustProcess("", config)
	}

	var secureOption otlptracegrpc.Option
	if config.Insecure {
		secureOption = otlptracegrpc.WithInsecure()
	} else {
		secureOption = otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	}

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(config.CollectorURL),
		),
	)
	if err != nil {
		panic(err)
	}

	otel.SetTracerProvider(
		sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(resource),
		),
	)
	return exporter.Shutdown
}

func NewTracingTransport() *otelhttp.Transport {
	return otelhttp.NewTransport(
		http.DefaultTransport,
		otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
			return otelhttptrace.NewClientTrace(ctx)
		}),
	)
}
