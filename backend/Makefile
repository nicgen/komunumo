SQLITE_DSN ?= ./komunumo.db
MIGRATIONS_DIR := internal/adapters/db/migrations

.PHONY: test test-domain test-application test-adapters test-http migrate-up migrate-down sqlc run lint tidy

test:
	go test ./...

test-domain:
	go test ./internal/domain/...

test-application:
	go test ./internal/application/...

test-adapters:
	go test ./internal/adapters/...

test-http:
	go test ./internal/adapters/http/...

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "sqlite://$(SQLITE_DSN)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "sqlite://$(SQLITE_DSN)" down 1

sqlc:
	sqlc generate

run:
	go run ./cmd/server

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
