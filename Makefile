NAME=pimpmypack
POSTGRES_CONTAINER=pimpmypack-postgres
POSTGRES_PORT=5432
POSTGRES_USER=pmp_user
POSTGRES_PASSWORD=pmp1234
POSTGRES_DB=pmp_db
POSTGRES_VERSION=16

.PHONY: test api-doc build lint start-db stop-db clean-db

start-db:
	@echo "Starting PostgreSQL container (version $(POSTGRES_VERSION))..."
	@docker run --name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		-d postgres:$(POSTGRES_VERSION) || true
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker exec $(POSTGRES_CONTAINER) pg_isready -h localhost -p 5432 -U $(POSTGRES_USER); do sleep 1; done
	@echo "Waiting additional 10 seconds for database to be fully ready..."
	@sleep 10

stop-db:
	@echo "Stopping PostgreSQL container..."
	@docker stop $(POSTGRES_CONTAINER) || true

clean-db: stop-db
	@echo "Removing PostgreSQL container..."
	@docker rm $(POSTGRES_CONTAINER) || true

test: start-db
	@echo "Running tests..."
	@go test -covermode=atomic -coverprofile=coverage.out -race ./...
	@$(MAKE) stop-db

api-doc:
	swag init --tags \!Internal

build: test
	go build

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout=5m 
