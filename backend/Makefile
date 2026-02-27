.PHONY: dev build test lint sqlc migrate-up migrate-down migrate-create setup

# Run the API server locally
dev:
	go run ./cmd/api/main.go

# Build the API binary
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api ./cmd/api/main.go

# Run all tests with coverage
test:
	go test ./... -v -cover -race

# Run integration tests
test-integration:
	go test ./... -v -tags=integration -race

# Lint with golangci-lint
lint:
	golangci-lint run ./...

# Database migrations
migrate-up:
	migrate -path internal/db/migrations -database "$${DATABASE_URL}" up

migrate-down:
	migrate -path internal/db/migrations -database "$${DATABASE_URL}" down 1

migrate-create:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

# SQLC code generation
sqlc:
	sqlc generate

# Full local setup
setup: migrate-up sqlc

# Format Go code
fmt:
	gofmt -s -w .

# Download dependencies
deps:
	go mod download && go mod tidy
