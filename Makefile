.PHONY: build build-pgkit test clean fmt vet tidy help localdb-start localdb-stop localdb-tables localdb-clean acceptance acceptance-dynamodbkit acceptance-pgkit localpostgres-start localpostgres-stop localpostgres-clean install-pgkit staticcheck

help:
	@echo "build                 - Build all packages"
	@echo "build-pgkit           - Build pgkit CLI with version info to bin/pgkit"
	@echo "test                  - Run tests (includes tidy, fmt, vet, staticcheck)"
	@echo "fmt                   - Format code"
	@echo "vet                   - Run go vet"
	@echo "staticcheck           - Run staticcheck"
	@echo "tidy                  - Tidy go modules"
	@echo "clean                 - Clean build artifacts"
	@echo "install-pgkit         - Install pgkit CLI tool to GOPATH/bin"
	@echo "localdb-start         - Start local DynamoDB and create tables"
	@echo "localdb-stop          - Stop local DynamoDB"
	@echo "localdb-clean         - Clean up local DynamoDB (stop containers and remove volumes)"
	@echo "localpostgres-start   - Start local PostgreSQL database (port 5433)"
	@echo "localpostgres-stop    - Stop local PostgreSQL database"
	@echo "localpostgres-clean   - Clean up local PostgreSQL (stop container and remove volume)"
	@echo "acceptance            - Run all acceptance tests"
	@echo "acceptance-dynamodbkit - Run DynamoDB acceptance tests"
	@echo "acceptance-pgkit      - Run pgkit acceptance tests"

build:
	go build ./...

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -ldflags "\
	-X main.version=$(VERSION) \
	-X main.gitCommit=$(GIT_COMMIT) \
	-X main.buildDate=$(BUILD_DATE)"

build-pgkit:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/pgkit ./cmd/pgkit

test: tidy fmt vet staticcheck
	go test -v ./... -tags=!acceptance

staticcheck:
	staticcheck ./...

fmt:
	gofmt -s -l -w .

vet:
	go vet ./...

staticcheck:
	staticcheck ./...

tidy:
	go mod tidy

clean:
	go clean
	rm -rf *.out bin

install-pgkit:
	go install ./cmd/pgkit

localdb-start:
	cd deployments/localdb && docker-compose up -d
	@echo "DynamoDB Local starting on http://localhost:8000"
	@echo "DynamoDB Admin starting on http://localhost:8001"
	@echo "Creating tables..."
	./deployments/localdb/create-tables.sh

localdb-stop:
	cd deployments/localdb && docker-compose down

localdb-tables:
	./deployments/localdb/create-tables.sh

localdb-clean:
	cd deployments/localdb && docker-compose down -v
	docker volume prune -f --filter label=com.docker.compose.project=localdb

localpostgres-start:
	docker run -d --name go-kit-postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_DB=testdb \
		-p 5433:5432 \
		postgres:16-alpine
	@echo "Waiting for PostgreSQL and testdb database to be ready..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do \
		if docker exec go-kit-postgres psql -U postgres -lqt 2>/dev/null | cut -d \| -f 1 | grep -qw testdb; then \
			echo "PostgreSQL is ready!"; \
			break; \
		fi; \
		sleep 1; \
	done

localpostgres-stop:
	docker stop go-kit-postgres || true
	docker rm go-kit-postgres || true

localpostgres-clean:
	docker stop go-kit-postgres || true
	docker rm go-kit-postgres || true
	docker volume prune -f

acceptance-dynamodbkit: localdb-start
	@AWS_ENDPOINT_URL=http://localhost:8000 \
	AWS_ACCESS_KEY_ID=dummy \
	AWS_SECRET_ACCESS_KEY=dummy \
	AWS_DEFAULT_REGION=us-east-1 \
	go test -count=1 -v ./test/acceptance/dynamodbkit -tags=acceptance; \
	test_result=$$?; \
	$(MAKE) localdb-stop; \
	$(MAKE) localdb-clean; \
	exit $$test_result

acceptance-pgkit: localpostgres-start
	@DATABASE_URL="postgres://postgres:postgres@localhost:5433/testdb?sslmode=disable" \
	go test -count=1 -v ./test/acceptance/pgkit -tags=acceptance; \
	test_result=$$?; \
	$(MAKE) localpostgres-stop; \
	$(MAKE) localpostgres-clean; \
	exit $$test_result

acceptance: acceptance-dynamodbkit acceptance-pgkit