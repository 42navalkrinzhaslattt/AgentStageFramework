# Emergent World Engine - Go Framework Makefile

# Build variables  
BINARY_DIR=bin

# Go build flags
LDFLAGS=-ldflags "-s -w"

# Default target
.PHONY: all
all: test example

# Build the framework example
.PHONY: example
example:
	@echo "Building framework example..."
	@mkdir -p $(BINARY_DIR)
	go build $(LDFLAGS) -o $(BINARY_DIR)/framework-example examples/simple/main.go

# Run the framework example
.PHONY: run-example
run-example:
	go run examples/simple/main.go

# Testing
.PHONY: test
test:
	go test -v ./...

.PHONY: test-race
test-race:
	go test -v -race ./...

.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Linting and formatting
.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint: fmt vet
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Dependency management
.PHONY: deps
deps:
	go mod tidy
	go mod download

.PHONY: deps-update
deps-update:
	go get -u ./...
	go mod tidy

# Clean targets
.PHONY: clean
clean:
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

# Help
.PHONY: help
help:
	@echo "Emergent World Engine - Go Framework"
	@echo "Available Make targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  all            - Test and build example"
	@echo "  example        - Build framework example"
	@echo "  run-example    - Run framework example"
	@echo ""
	@echo "Testing targets:"
	@echo "  test           - Run framework tests"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo ""
	@echo "Quality targets:"
	@echo "  fmt            - Format code"
	@echo "  vet            - Run go vet"
	@echo "  lint           - Run linter"
	@echo ""
	@echo "Dependency targets:"
	@echo "  deps           - Tidy and download dependencies"
	@echo "  deps-update    - Update all dependencies"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean          - Clean build artifacts"
	@echo "  help           - Show this help message"