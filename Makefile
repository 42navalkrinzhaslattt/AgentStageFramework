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

# Presidential Simulator (game submodule)
GAME_DIR=game
GAME_BIN=$(BINARY_DIR)/presidential-sim
ENV_FILE?=.env

# Load .env (export lines) if present
ifneq (,$(wildcard $(ENV_FILE)))
include $(ENV_FILE)
export $(shell sed -n 's/^[A-Za-z_][A-Za-z0-9_]*=//p' $(ENV_FILE) >/dev/null 2>&1)
endif

.PHONY: game-build
game-build:
	@echo "Building Presidential Simulator..."
	@mkdir -p $(BINARY_DIR)
	cd $(GAME_DIR) && go build $(LDFLAGS) -o ../$(GAME_BIN) .

.PHONY: game-run
game-run: game-build
	@echo "Running Presidential Simulator (terminal mode)..."
	./$(GAME_BIN)

.PHONY: game-run-web
game-run-web: game-build
	@echo "Running Presidential Simulator (web mode on PORT=$(PORT))..."
	@PORT?=8080
	THETA_API_KEY=$${THETA_API_KEY} ./$(GAME_BIN) web $${PORT}

.PHONY: game-dev
game-dev:
	@echo "Running Presidential Simulator live (go run) with auto .env load..."
	cd $(GAME_DIR) && go run . web $${PORT:-8080}

.PHONY: game-clean
game-clean:
	rm -f $(GAME_BIN)

# Update help
help: ## show help
	@echo "Game targets:"; \
	echo "  game-build     - Build presidential simulator binary"; \
	echo "  game-run       - Run in terminal mode"; \
	echo "  game-run-web   - Run web server (PORT env)"; \
	echo "  game-dev       - Run via go run with live changes"; \
	echo "  game-clean     - Remove built binary";