.PHONY: build run clean test lint docker-up docker-down db-up db-down redis-up redis-down

# Configuration
BINARY_NAME=shortlink-go
APP_DIR=./cmd/shortlink-go

# Build the Go binary
build:
	go build -o $(BINARY_NAME) $(APP_DIR)

# Run the application
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

lint:
	docker run -t --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.57.2 golangci-lint run -v ./cmd/shortlink-go/...

fmt:
	go fmt ./...

# Generate Swagger documentation
swagger:
	swag init -d ./cmd/shortlink-go/

### Database management

DB_USER=postgres
DB_CONTAINER_NAME=shortlink-postgres
DB_PASSWORD=password
DB_PORT=5432
DB_NAME=db

# Start PostgreSQL container
db-up:
	docker run --name $(DB_CONTAINER_NAME) -e POSTGRES_USER=${DB_USER} -e POSTGRES_PASSWORD=$(DB_PASSWORD) -e POSTGRES_DB=$(DB_NAME) -p $(DB_PORT):5432 -d postgres

# Stop and remove the PostgreSQL container
db-down:
	docker stop $(DB_CONTAINER_NAME) && docker rm $(DB_CONTAINER_NAME)

# Database migration
MIGRATE_DOCKER_IMAGE=migrate/migrate
MIGRATION_PATH=./db/migrations
DATABASE_URL=postgres://${DB_USER}:${DB_PASSWORD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable
db-migrate:
	docker run --network host --rm -v ${PWD}/${MIGRATION_PATH}:/migrations ${MIGRATE_DOCKER_IMAGE} -path=/migrations/ -database "${DATABASE_URL}" up

### Redis

REDIS_CONTAINER_NAME=shortlink-redis
REDIS_PORT=6379

# Start Redis container
redis-up:
	docker run --name $(REDIS_CONTAINER_NAME) -p $(REDIS_PORT):6379 -d redis

# Stop and remove the Redis container
redis-down:
	docker stop $(REDIS_CONTAINER_NAME) && docker rm $(REDIS_CONTAINER_NAME)
