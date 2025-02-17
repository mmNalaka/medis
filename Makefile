# Run the main go program
.PHONY: run
run:
	go run cmd/medis/main.go || true

.PHONY: build test clean coverage coverage-html

# Build the application
build:
	go build -o bin/medis cmd/medis/main.go

# Run all tests
test:
	go test -v ./...

# Generate test coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Generate HTML coverage report
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

# Run the server
run: build
	./bin/medis
