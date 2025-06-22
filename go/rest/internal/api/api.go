//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yml ../../cmd/server/static/api.yml
package api

import (
	"net/http"

	"go-rest/internal/config"

	"github.com/rs/zerolog"
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
	WriteJson(w, api.config)
}

func (api *APIServer) GetCrash(w http.ResponseWriter, r *http.Request) {
	panic("Intentional panic for testing!")
}
