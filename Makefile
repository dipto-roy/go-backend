.PHONY: run build test lint migrate-up migrate-down swagger tidy

run:
	go run ./cmd/api

build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/api ./cmd/api

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

migrate-up:
	migrate -path migrations -database "$$DB_DSN" up

migrate-down:
	migrate -path migrations -database "$$DB_DSN" down 1

swagger:
	~/go/bin/swag init -g cmd/api/main.go -o docs

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

.DEFAULT_GOAL := run
