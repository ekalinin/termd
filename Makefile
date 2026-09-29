BINARY      := termd
PKG         := ./cmd/termd
GOLDEN_PKGS := ./internal/render ./internal/diagram ./internal/highlight ./internal/site
FILE        ?= testdata/regression.md
SITE_DIR    := _site

.DEFAULT_GOAL := help

.PHONY: help build install run test vet fmt fmt-check golden tidy check clean site

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  %-10s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary into ./termd
	go build -o $(BINARY) $(PKG)

install: ## Install termd into GOBIN
	go install $(PKG)

run: ## Render FILE with termd (default: testdata/regression.md)
	go run $(PKG) $(FILE)

test: ## Run all tests
	go test ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format the code with gofmt
	gofmt -w cmd internal

fmt-check: ## Fail if any file needs gofmt
	@out=$$(gofmt -l cmd internal); if [ -n "$$out" ]; then echo "needs gofmt:"; echo "$$out"; exit 1; fi

golden: ## Regenerate golden files after an intended output change
	go test $(GOLDEN_PKGS) -update

tidy: ## Tidy go.mod and go.sum
	go mod tidy

check: fmt-check vet test ## Run the format check, vet and tests

site: ## Build the landing page into ./_site
	go run ./cmd/termd-site $(SITE_DIR)

clean: ## Remove the built binary
	rm -f $(BINARY)
