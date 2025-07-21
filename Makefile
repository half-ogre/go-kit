.PHONY: build test clean fmt vet

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