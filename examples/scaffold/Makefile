.PHONY: help build run test test-integration lint generate config-md docker

SVC         := <svc>
VERSION     := $(shell cat VERSION)
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
GIT_BRANCH  := $(shell git symbolic-ref --short -q HEAD 2>/dev/null || echo unknown)
LDFLAGS     := -X github.com/bldsoft/gost/version.Version=$(VERSION) \
               -X github.com/bldsoft/gost/version.GitCommit=$(GIT_COMMIT) \
               -X github.com/bldsoft/gost/version.GitBranch=$(GIT_BRANCH)

# Local defaults for `make run`; every variable is documented in config.md.
export ADMIN_API_KEY   ?= dev-key
export DATABASE_SERVER ?= mongodb://127.0.0.1:27017/?replicaSet=rs0&directConnection=true
export DATABASE_DBNAME ?= $(SVC)

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-18s\033[0m %s\n", $$1, $$2}'
build: ## Build bin/$(SVC) with the version baked in
	go build -ldflags "$(LDFLAGS)" -o bin/$(SVC) ./cmd/$(SVC)
run: build ## Build and run against a local MongoDB replica set
	./bin/$(SVC)
test: ## Unit and route tests (no database needed)
	go test -race -count=1 ./...
test-integration: ## Integration tests; skipped locally when MongoDB is unreachable, fatal in CI
	go test -race -count=1 -tags=integration_test ./...
lint: ## golangci-lint (same rules as gost)
	golangci-lint run ./...
generate: ## Regenerate mocks and enums
	go generate ./...
config-md: ## Regenerate config.md from the Config struct
	UPDATE_CONFIG_MD=1 go test -count=1 ./pkg/config -run TestConfigDescriptionGolden
docker: ## Build the image
	docker build -t $(SVC):dev --build-arg COMMIT=$(GIT_COMMIT) --build-arg BRANCH=$(GIT_BRANCH) .
