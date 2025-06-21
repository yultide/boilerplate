.PHONY: __default help build
__default: help

version=0.1.0

build: internal/api/gen.go ## Build server
	go build -ldflags "-X go-rest/internal/config.version=$(version)" $(args) -o server cmd/server/main.go

run: internal/api/gen.go ## Run server
	go tool gow run cmd/server/main.go $(args)

help: ## Show this help
	@grep -E '^[a-zA-Z-][a-zA-Z_-]+:.*?' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":[^#]*(##)?"}; {printf "\033[34m%-25s\033[0m %s\n", $$1, $$2}'

### dependencies
internal/api/gen.go: api/api.yml
	go generate internal/api/api.go

