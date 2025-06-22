package log

import (
	"context"
	"encoding/json"
	"fmt"
	"go-rest/internal/config"
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
	serviceName = "go-rest"
)

// tracerProvider returns an OpenTelemetry TracerProvider configured to export to Jaeger.
func tracerProvider(ctx context.Context, cfg *config.Config) (*tracesdk.TracerProvider, error) {
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(cfg.OtelEndpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create jaeger exporter: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL, // Use OpenTelemetry semantic conventions schema URL
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(cfg.Version),
			attribute.String("environment", "development"),
		)),
	)
	return tp, nil
}

var once sync.Once

func OpenTelemetryMiddlware(r chi.Routes, cfg *config.Config) func(next http.Handler) http.Handler {
	once.Do(func() {
		tp, _ := tracerProvider(context.Background(), cfg)
		otel.SetTracerProvider(tp)
	})

	return otelchi.Middleware(serviceName,
		otelchi.WithRequestMethodInSpanName(true),
		otelchi.WithChiRoutes(r))
}

func toAttrs(m map[string]interface{}) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(m))
	for k, i := range m {
		switch v := i.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, v))
		case int:
			attrs = append(attrs, attribute.Int(k, v))
		case float64:
			attrs = append(attrs, attribute.Float64(k, v))
		case []string:
			attrs = append(attrs, attribute.StringSlice(k, v))
		default:
			jv, _ := json.Marshal(v)
			attrs = append(attrs, attribute.String(k, string(jv)))
		}
	}
	return attrs
}
