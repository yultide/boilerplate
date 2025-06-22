package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"
	"strconv"
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

var hook = zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
	logData := map[string]interface{}{}
	ev := fmt.Sprintf("%s}", reflect.ValueOf(e).Elem().FieldByName("buf"))
	json.Unmarshal([]byte(ev), &logData)
	span := trace.SpanFromContext(e.GetCtx())
	span.AddEvent(msg, trace.WithAttributes(
		toAttrs(logData)...,
	), trace.WithStackTrace(true))
})

func LoggerMiddleware(logger *zerolog.Logger, cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			span := trace.SpanFromContext(ctx)
			reqId := middleware.GetReqID(ctx)

			log := logger.With().Ctx(ctx).Fields(map[string]interface{}{
				"req-id": reqId,
			}).
				Logger().
				Hook(hook)

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
				bytesIn, _ := strconv.Atoi(r.Header.Get("Content-Length"))
				fields := map[string]interface{}{
					"remote-ip":  r.RemoteAddr,
					"url":        r.URL.Path,
					"proto":      r.Proto,
					"method":     r.Method,
					"headers":    toMap(r.Header),
					"user-agent": r.Header.Get("User-Agent"),
					"status":     ww.Status(),
					"latency-ms": float64(t2.Sub(t1).Nanoseconds()) / 1000000.0,
					"bytes-in":   bytesIn,
					"bytes-out":  ww.BytesWritten(),
				}
				if body != nil {
					fields["body"] = body
				}

				log.Info().
					Fields(fields).
					Msgf("%s %s", r.Method, r.URL.Path)

				span.SetAttributes(toAttrs(fields)...)
			}()

			// save log to context
			rr := r.WithContext(log.WithContext(r.Context()))

			next.ServeHTTP(ww, rr)
		}
		return http.HandlerFunc(fn)
	}
}

func toMap(h http.Header) map[string]interface{} {
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
