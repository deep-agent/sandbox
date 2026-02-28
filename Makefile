.PHONY: build build-linux clean run docker-build docker-run test ensure-env

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

docker-build:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE_NAME) -f docker/Dockerfile --push $(DOCKER_BUILD_ARGS) .

docker-reload: ensure-env
	$(DOCKER_ENV) docker compose down && $(DOCKER_ENV) docker compose up -d

docker-rebuild: ensure-env
	$(DOCKER_ENV) docker compose up --build --force-recreate -d


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
