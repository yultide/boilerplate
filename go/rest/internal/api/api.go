//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yml ../../cmd/server/static/api.yml
package api

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go-rest/internal/config"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type APIServer struct {
	config *config.Config
}

func NewServer(config *config.Config) *APIServer {
	return &APIServer{
		config: config,
	}
}

func (api *APIServer) GetHealth(w http.ResponseWriter, r *http.Request) {
	p := map[string]string{
		"status": "ok",
	}
	log := zerolog.Ctx(r.Context())
	log.Info().Msg("Healthy!!!")
	WriteJson(w, p)
}

func (api *APIServer) GetConfig(w http.ResponseWriter, r *http.Request) {
	targetURL := "https://httpbin.org/delay/0.5"
	fetchData(r.Context(), targetURL)
	fetchData(r.Context(), targetURL)

	WriteJson(w, api.config)
}

func (api *APIServer) GetCrash(w http.ResponseWriter, r *http.Request) {
	panic("Intentional panic for testing!")
}

func fetchData(ctx context.Context, url string) (string, error) {
	client := otelhttp.DefaultClient

	// Create a new HTTP GET request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Perform the HTTP request.
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	defer resp.Body.Close() // Ensure the response body is closed

	// Check if the response status code indicates an error.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	// Read the response body.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}
