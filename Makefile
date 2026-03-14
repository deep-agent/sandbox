.PHONY: build build-linux clean run push-to-dockerhub docker-run test ensure-env

BINARY_NAME=sandbox
IMAGE_NAME=fanlv/sandbox:latest
GO=go

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
	$(GO) run ./cmd/sandbox-server

run-mcp:
	$(GO) run ./cmd/mcp-hub

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
