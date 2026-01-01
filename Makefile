.PHONY: build test integration-test clean docker-build

# Build the client binary
build:
	go build -o go-send-client cmd/client/main.go

# Run unit tests
test:
	go test -v ./...

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
