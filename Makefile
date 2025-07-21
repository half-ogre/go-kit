.PHONY: build test clean fmt vet help

help:
	@echo "build  - Build all packages"
	@echo "test   - Run tests"
	@echo "fmt    - Format code"
	@echo "vet    - Run go vet"
	@echo "clean  - Clean build artifacts"

build:
	go build ./...

test:
	go test -v ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	go clean
	rm -rf *.out