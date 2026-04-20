.PHONY: all fmt lint test build tidy check clean

all: check fmt lint test build

fmt:
	go fmt ./...

lint:
	golangci-lint run ./...

test:
	go test -v ./...

build:
	go build ./...

tidy:
	go mod tidy
	go mod verify

check: tidy
	go vet ./...

clean:
	rm -f coverage.out coverage.html
	go clean -testcache
