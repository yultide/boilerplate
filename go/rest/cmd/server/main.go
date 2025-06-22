package main

import (
	"embed"
	"fmt"
	"net/http"

	"go-rest/internal/api"
	"go-rest/internal/config"
	"go-rest/internal/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed static
var staticFs embed.FS

func main() {
	cfg := config.LoadConfig()
	logger := log.NewLogger()

	// create a type that satisfies the `api.ServerInterface`, which contains an implementation of every operation from the generated code
	server := api.NewServer(cfg)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(log.OpenTelemetryMiddlware(r, cfg))
	r.Use(log.LoggerMiddleware(logger, cfg))

	// get an `http.Handler` that we can use
	h := api.HandlerFromMux(server, r)

	// swagger ui
	r.Handle("GET /*", api.StaticSite(staticFs))

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	s := &http.Server{
		Handler: h,
		Addr:    addr,
	}

	logger.Info().Msgf("Version %s", cfg.Version)
	logger.Info().Msgf("Server started on %s", addr)
	// And we serve HTTP until the world ends.
	s.ListenAndServe()
}
