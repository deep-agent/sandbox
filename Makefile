.PHONY: build build-linux clean run docker-build docker-run test

BINARY_NAME=sandbox
IMAGE_NAME=fanlv/sandbox:latest
GO=go

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
	docker buildx build --platform linux/amd64,linux/arm64 -t $(IMAGE_NAME) -f docker/Dockerfile --push .

docker-reload:
	docker compose down && docker compose up -d

docker-rebuild:
	docker compose up --build --force-recreate -d


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
