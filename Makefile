.PHONY: build build-linux clean run run-all stop-all tail-log push-to-dockerhub docker-run test ensure-env cdp cdp-stop

BINARY_NAME=sandbox
IMAGE_NAME=fanlv/sandbox:latest
GO=go

WORKSPACE ?= ${LOCAL_MEMORY}/workspace
SANDBOX_SRV_PORT ?= 8000
MCP_HUB_PORT ?= 8001
JWT_SECRET ?=
JWT_AUTH_REQUIRED ?= false
LOG_FILE ?= /tmp/sandbox.log

HTTP_PROXY ?=
HTTPS_PROXY ?=
NO_PROXY ?= localhost,127.0.0.1

DOCKER_BUILD_ARGS :=
DOCKER_ENV :=
ifneq ($(HTTP_PROXY),)
	DOCKER_BUILD_ARGS += --build-arg HTTP_PROXY=$(HTTP_PROXY)
	DOCKER_ENV += HTTP_PROXY=$(HTTP_PROXY)
endif
ifneq ($(HTTPS_PROXY),)
	DOCKER_BUILD_ARGS += --build-arg HTTPS_PROXY=$(HTTPS_PROXY)
	DOCKER_ENV += HTTPS_PROXY=$(HTTPS_PROXY)
endif
ifneq ($(NO_PROXY),)
	DOCKER_BUILD_ARGS += --build-arg NO_PROXY=$(NO_PROXY)
	DOCKER_ENV += NO_PROXY=$(NO_PROXY)
endif

ensure-env:
	@if [ ! -f .env ]; then \
		echo "Creating .env from .env.example..."; \
		cp .env.example .env; \
	fi

build:
	$(GO) build -o bin/sandbox-server ./cmd/sandbox-server
	$(GO) build -o bin/mcp-hub ./cmd/mcp-hub

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o bin/sandbox-server ./cmd/sandbox-server
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -o bin/mcp-hub ./cmd/mcp-hub

clean:
	rm -rf bin/

run-server:
	env WORKSPACE="$(WORKSPACE)" \
		SANDBOX_SRV_PORT="$(SANDBOX_SRV_PORT)" \
		JWT_SECRET="$(JWT_SECRET)" \
		JWT_AUTH_REQUIRED="$(JWT_AUTH_REQUIRED)" \
		$(GO) run ./cmd/sandbox-server

run-mcp:
	env WORKSPACE="$(WORKSPACE)" \
		MCP_HUB_PORT="$(MCP_HUB_PORT)" \
		$(GO) run ./cmd/mcp-hub

run-all: build
	@: > $(LOG_FILE)
	@echo "Starting sandbox-server and mcp-hub..."
	@nohup sh -c 'env WORKSPACE="$(WORKSPACE)" SANDBOX_SRV_PORT="$(SANDBOX_SRV_PORT)" JWT_SECRET="$(JWT_SECRET)" JWT_AUTH_REQUIRED="$(JWT_AUTH_REQUIRED)" ./bin/sandbox-server 2>&1 | sed -u "s/^/[server]  /" >> $(LOG_FILE)' >/dev/null 2>&1 &
	@nohup sh -c 'env WORKSPACE="$(WORKSPACE)" MCP_HUB_PORT="$(MCP_HUB_PORT)" ./bin/mcp-hub 2>&1 | sed -u "s/^/[mcp-hub] /" >> $(LOG_FILE)' >/dev/null 2>&1 &
	@sleep 1
	@echo "Logs  -> $(LOG_FILE)"
	@echo "Stop  -> make stop-all"
	@echo "---- tailing log (Ctrl+C to exit, services keep running) ----"
	@tail -f $(LOG_FILE)

stop-all:
	@pkill -f 'bin/sandbox-server' 2>/dev/null || true
	@pkill -f 'bin/mcp-hub' 2>/dev/null || true
	@echo "Stopped sandbox-server and mcp-hub"

tail-log:
	@tail -f $(LOG_FILE)

push-to-dockerhub:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE_NAME) -f docker/Dockerfile --push $(DOCKER_BUILD_ARGS) .

docker-start: ensure-env
	$(DOCKER_ENV) docker compose up -d

docker-restart: ensure-env
	$(DOCKER_ENV) docker compose down && $(DOCKER_ENV) docker compose up -d

docker-dev: ensure-env
	$(DOCKER_ENV) docker compose -f docker-compose.debug.yaml up --build --force-recreate -d

docker-restart-dev: ensure-env
	$(DOCKER_ENV) docker compose -f docker-compose.debug.yaml down && $(DOCKER_ENV) docker compose -f docker-compose.debug.yaml up -d

docker-down: ensure-env
	$(DOCKER_ENV) docker compose down

docker-down-dev: ensure-env
	$(DOCKER_ENV) docker compose -f docker-compose.debug.yaml down

nginx-reload:
	docker exec sandbox nginx -s reload

nginx-test:
	docker exec sandbox nginx -t

test:
	$(GO) test -v ./...

tidy:
	$(GO) mod tidy

fmt:
	$(GO) fmt ./...

lint:
	golangci-lint run

