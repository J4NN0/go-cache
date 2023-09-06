# === CONFIG =======================================================
COVER_PROFILE="cover.out"


# === TEST =======================================================
test:
	@echo "---> Running all tests"
	go test -race -cover -coverprofile=$(COVER_PROFILE) ./...
.PHONY: test


# === TOOLS =======================================================
# Get a decorated HTML presentation of cover file: showing the covered (green), uncovered (red), and un-instrumented (grey) source.
tool-cover:
	go tool cover -html=$(COVER_PROFILE)
.PHONY: tool-cover

# Fix go.mod and go.sum
tool-tidy:
	@echo "---> Checking module requirements"
	go mod tidy
.PHONY: tool-tidy

# Format go code
tool-fmt:
	@echo "---> Formatting code"
	go fmt ./...
.PHONY: tool-fmt

# Examine Go source code and reports suspicious constructs
tool-vet:
	@echo "---> Checking Go source code"
	go vet ./...
.PHONY: tool-vet

# Run application using linters: it runs linters in parallel, uses caching, supports yaml config, etc.
tool-lint:
	@echo "---> Running linter"
	golangci-lint run ./... --timeout=3m
.PHONY: tool-lint


# === DEVELOPMENT =======================================================
pre-commit: tool-tidy tool-fmt tool-vet tool-lint test
