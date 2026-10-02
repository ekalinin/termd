# Build and test the greet command.
BINARY := greet

.PHONY: build test

build: ## Build the binary
	go build -o $(BINARY) .

# Tests run with the race detector.
test:
	go test -race ./...
