.PHONY: build run clean test

# Build the binary
build:
	go build -o dojo ./cmd/dojo

# Build and run
run: build
	./dojo

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f dojo
