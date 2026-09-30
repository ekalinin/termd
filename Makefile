BINARY      := termd
PKG         := ./cmd/termd
GOLDEN_PKGS := ./internal/render ./internal/diagram ./internal/highlight ./internal/site
FILE        ?= testdata/regression.md
SITE_DIR    := _site

.DEFAULT_GOAL := help

.PHONY: help build install run test vet vet-windows fmt fmt-check golden tidy tidy-check check clean site release-check release-snapshot

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

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

vet-windows: ## Run go vet for the Windows build
	GOOS=windows go vet ./...

fmt: ## Format the code with gofmt
	gofmt -w cmd internal

fmt-check: ## Fail if any file needs gofmt
	@out=$$(gofmt -l cmd internal); if [ -n "$$out" ]; then echo "needs gofmt:"; echo "$$out"; exit 1; fi

golden: ## Regenerate golden files after an intended output change
	go test $(GOLDEN_PKGS) -update

tidy: ## Tidy go.mod and go.sum
	go mod tidy

tidy-check: ## Fail if go.mod or go.sum needs go mod tidy
	go mod tidy -diff

check: fmt-check tidy-check vet vet-windows test ## Run the format and tidy checks, vet (host and Windows) and tests

site: ## Build the landing page into ./_site
	go run ./cmd/termd-site $(SITE_DIR)

release-check: ## Validate the GoReleaser config
	goreleaser check

release-snapshot: ## Build the release archives into ./dist without publishing
	goreleaser release --snapshot --clean

clean: ## Remove the built binary
	rm -f $(BINARY)
