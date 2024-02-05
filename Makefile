NAME=pimpmypack

test:
	go test -covermode=atomic -coverprofile=coverage.out -race ./...

api-doc:
	swag init 

build: test
	go build

lint:
	@echo "Running staticcheck..."
	@staticcheck ./...
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout=5m --out-format=colored-line-number	