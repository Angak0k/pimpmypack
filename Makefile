NAME=pimpmypack

test:
	go test -covermode=atomic -coverprofile=coverage.out -race ./...

api-doc:
	swag init --tags \!Internal

build: test
	go build

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout=5m 
