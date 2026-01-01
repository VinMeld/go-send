.PHONY: build test integration-test clean docker-build

# Build the client binary
build:
	go build -o go-send-client cmd/client/main.go

# Run unit tests
test:
	GOTOOLCHAIN=go1.25.5+auto go test -v ./...

# Run unit tests with coverage (excludes cmd packages)
coverage:
	GOTOOLCHAIN=go1.25.5+auto go test -v -coverprofile=coverage.out $$(go list ./... | grep -v cmd)
	go tool cover -func=coverage.out

# Run integration tests
integration-test: build
	./tests/integration/run.sh

# Build Docker image
docker-build:
	docker build -t go-send-server:latest .

# Clean up binaries and artifacts
clean:
	rm -f go-send-client
	rm -f coverage.out
	rm -rf test_data_*
