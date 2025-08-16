.PHONY: help run build test clean docker-up docker-down migrate

help:
	@echo "Available commands:"
	@echo "  make run         - Run the application locally"
	@echo "  make build       - Build the application"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker-up   - Start with Docker Compose"
	@echo "  make docker-down - Stop Docker Compose"
	@echo "  make migrate     - Run database migrations"

run:
	go run cmd/server/main.go

build:
	go build -o bin/websocket-server cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

migrate:
	@echo "Migrations are run automatically on startup"

# Development commands
dev:
	air -c .air.toml

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run