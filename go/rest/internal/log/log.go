package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"go-rest/internal/config"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func NewLogger() *zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &logger
}

func LoggerMiddleware(logger *zerolog.Logger, cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			span := trace.SpanFromContext(ctx)
			reqId := middleware.GetReqID(ctx)
			log := logger.With().Fields(map[string]interface{}{
				"req-id": reqId,
			}).Logger()
			span.SetAttributes(attribute.String("req-id", middleware.GetReqID(ctx)))

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxRequestBody)

			var body interface{}

			// Only log for POST requests with application/json content type
			if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {
				// Read the request body
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					msg := fmt.Sprintf("failed to read request body: %s", err.Error())
					log.Error().Err(err).Msg(msg)
					http.Error(w, msg, http.StatusInternalServerError)
					return
				}

				// Restore the request body for the next handler
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				err = json.Unmarshal(bodyBytes, &body)
				if err != nil {
					log.Error().Err(err).Msgf("failed to parse request body: %s", err.Error())
					http.Error(w, "Failed to parse request body", http.StatusBadRequest)
					return
				}
			}

			t1 := time.Now()
			defer func() {
				t2 := time.Now()

				// Recover and record stack traces in case of a panic
				if rec := recover(); rec != nil {
					log.Error().
						Interface("recover_info", rec).
						Bytes("debug_stack", debug.Stack()).
						Msg("log system error")
					http.Error(ww, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}

				// log end request
				fields := map[string]interface{}{
					"remote-ip":  r.RemoteAddr,
					"url":        r.URL.Path,
					"proto":      r.Proto,
					"method":     r.Method,
					"headers":    ToMap(r.Header),
					"body":       body,
					"user-agent": r.Header.Get("User-Agent"),
					"status":     ww.Status(),
					"latency-ms": float64(t2.Sub(t1).Nanoseconds()) / 1000000.0,
					"bytes-in":   r.Header.Get("Content-Length"),
					"bytes-out":  ww.BytesWritten(),
				}
				log.Info().
					Fields(fields).
					Msgf("%s %s", r.Method, r.URL.Path)

				for k, i := range fields {
					switch v := i.(type) {
					case string:
						span.SetAttributes(attribute.String(k, v))
					case int:
						span.SetAttributes(attribute.Int(k, v))
					case float64:
						span.SetAttributes(attribute.Float64(k, v))
					case []string:
						span.SetAttributes(attribute.StringSlice(k, v))
					default:
						jv, _ := json.Marshal(v)
						span.SetAttributes(attribute.String(k, string(jv)))
					}
				}
			}()

			// save log to context
			rr := r.WithContext(log.WithContext(r.Context()))

			next.ServeHTTP(ww, rr)
		}
		return http.HandlerFunc(fn)
	}
}

func ToMap(h http.Header) map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range h {
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			m[k] = v
		}
	}
	return m
}
