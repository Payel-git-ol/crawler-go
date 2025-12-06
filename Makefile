.PHONY: help build run clean test deps docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make test        - Run tests"
	@echo "  make deps        - Download dependencies"
	@echo "  make docker-up   - Start Docker services (Typesense)"
	@echo "  make docker-down - Stop Docker services"

build:
	@echo "Building Fyne-on..."
	go build -o bin/app.exe ./cmd/app

run: build
	@echo "Running Fyne-on..."
	@echo "API will be available at http://localhost:3000"
	@echo "API Routes: http://localhost:3000/api/routes"
	./bin/app.exe

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf badger_data/

test:
	@echo "Running tests..."
	go test -v ./...

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Linting code..."
	golangci-lint run ./...

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs:
	docker-compose logs -f

init:
	@echo "Initializing project..."
	mkdir -p bin badger_data

.DEFAULT_GOAL := help
