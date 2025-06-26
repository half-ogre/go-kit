.PHONY: build test clean all

all: build test

build:
	go build ./...

test:
	go test -v ./...

clean:
	go clean
	rm -rf *.out