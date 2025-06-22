package log

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0" // Use a specific version for semantic conventions
)

const (
	serviceName    = "chi-go-service"
	serviceVersion = "1.0.0"
	jaegerURL      = "http://localhost:4318/v1/traces" // OLTP gRPC
)

// tracerProvider returns an OpenTelemetry TracerProvider configured to export to Jaeger.
func tracerProvider(ctx context.Context) (*tracesdk.TracerProvider, error) {
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpointURL(jaegerURL))
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
	}

	// Create a new TracerProvider.
	// Always set the TracerProvider, so that tracing can be enabled.
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate. AlwaysSample samples all traces.
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		// Add the Jaeger exporter.
		tracesdk.WithBatcher(exp),
		// Define the service resource. This resource describes the entity producing the telemetry.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL, // Use OpenTelemetry semantic conventions schema URL
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			attribute.String("environment", "development"),
		)),
	)
	return tp, nil
}

var once sync.Once

func OpenTelemetryMiddlware(r chi.Routes) func(next http.Handler) http.Handler {
	once.Do(func() {
		tp, _ := tracerProvider(context.Background())
		otel.SetTracerProvider(tp)
	})

	return otelchi.Middleware(serviceName,
		otelchi.WithRequestMethodInSpanName(true),
		otelchi.WithChiRoutes(r))
}
