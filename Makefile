NAME=pimpmypack

test:
	go test -covermode=atomic -coverprofile=coverage.out -race ./...

api-doc:
	swag init 

build: test
	go build

lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v1.52 golangci-lint run -v
