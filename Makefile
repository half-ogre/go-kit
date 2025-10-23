.PHONY: build test clean fmt vet tidy help localdb-start localdb-stop localdb-tables localdb-clean acceptance

help:
	@echo "build          - Build all packages"
	@echo "test           - Run tests (includes tidy, fmt, vet)"
	@echo "fmt            - Format code"
	@echo "vet            - Run go vet"
	@echo "tidy           - Tidy go modules"
	@echo "clean          - Clean build artifacts"
	@echo "localdb-start  - Start local DynamoDB and create tables"
	@echo "localdb-stop   - Stop local DynamoDB"
	@echo "localdb-clean  - Clean up local DynamoDB (stop containers and remove volumes)"
	@echo "acceptance     - Run acceptance tests (starts/stops/cleans local DynamoDB automatically)"

build:
	go build ./...

test: tidy fmt vet staticcheck
	go test -v ./... -tags=!acceptance

staticcheck:
	staticcheck ./...

fmt:
	gofmt -s -l -w .

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	go clean
	rm -rf *.out

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

acceptance: localdb-start
	@AWS_ENDPOINT_URL=http://localhost:8000 \
	AWS_ACCESS_KEY_ID=dummy \
	AWS_SECRET_ACCESS_KEY=dummy \
	AWS_DEFAULT_REGION=us-east-1 \
	go test -count=1 -v ./test/acceptance/dynamodbkit -tags=acceptance; \
	test_result=$$?; \
	$(MAKE) localdb-stop; \
	$(MAKE) localdb-clean; \
	exit $$test_result