NAME=pimpmypack
POSTGRES_CONTAINER=pimpmypack-postgres
POSTGRES_PORT=5432
POSTGRES_USER=pmp_user
POSTGRES_PASSWORD=pmp1234
POSTGRES_DB=pmp_db
POSTGRES_VERSION=17

.PHONY: test api-doc build lint start-db stop-db clean-db build-apitest api-test api-test-auth stop-server start-server restart-server api-test-full

start-db:
	@echo "Starting PostgreSQL container (version $(POSTGRES_VERSION))..."
	@docker rm -fv $(POSTGRES_CONTAINER) 2>/dev/null || true
	@docker run --name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		-d postgres:$(POSTGRES_VERSION)
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker exec $(POSTGRES_CONTAINER) pg_isready -h localhost -p 5432 -U $(POSTGRES_USER); do sleep 1; done
	@echo "Waiting additional 10 seconds for database to be fully ready..."
	@sleep 10

stop-db:
	@echo "Stopping PostgreSQL container..."
	@docker stop $(POSTGRES_CONTAINER) || true

clean-db: stop-db
	@echo "Removing PostgreSQL container..."
	@docker rm -v $(POSTGRES_CONTAINER) || true

test: start-db
	@echo "Running tests..."
	@go test -covermode=atomic -coverprofile=coverage.out -race -p=1 ./...
	@$(MAKE) clean-db

api-doc:
	swag init --output api-doc --generalInfo main.go --tags \!Internal

build:
	go build

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout=5m

build-apitest:
	@echo "Building API test CLI..."
	@go build -o bin/apitest ./tests/cmd/apitest

api-test: build-apitest
	@echo "Running API tests with Go CLI..."
	@echo "NOTE: Server must be running (STAGE=LOCAL)"
	@./bin/apitest run --all

stop-server:
	@echo "Stopping any running pimpmypack server..."
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@pkill -f "go run main.go" || true
	@pkill -f "./$(NAME)" || true
	@pkill -f "$(NAME)" || true
	@sleep 2

start-server: stop-server clean-db start-db
	@echo "Starting pimpmypack server in background..."
	@nohup go run main.go > /tmp/pimpmypack-server.log 2>&1 &
	@echo "Waiting for server to start..."
	@sleep 3
	@echo "Server started. Logs: /tmp/pimpmypack-server.log"

restart-server: start-server

api-test-full: start-server build-apitest
	@echo "Running API tests with fresh server..."
	@./bin/apitest run --all || (echo "Tests failed. Server logs:"; tail -50 /tmp/pimpmypack-server.log; exit 1)
	@echo ""
	@echo "✅ Tests completed successfully"
	@echo "Server is still running. Use 'make stop-server' to stop it."
