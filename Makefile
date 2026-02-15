.PHONY: run build test clean docker-up docker-down migrate-up migrate-down

# Build the application
build:
	go build -o bin/eraflazz cmd/api/main.go

# Run the application locally
run:
	go run cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Database migrations
migrate-up:
	migrate -path migrations -database "postgres://eraflazz_user:eraflazz_password@localhost:5432/eraflazz_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://eraflazz_user:eraflazz_password@localhost:5432/eraflazz_db?sslmode=disable" down

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Development setup
dev-setup:
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Lint
lint:
	golangci-lint run
