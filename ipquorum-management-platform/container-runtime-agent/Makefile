.PHONY: help build test clean run docker-build docker-run lint fmt vet

# Variables
BINARY_NAME=ipquorum-agent
VERSION?=dev
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}"

# Docker variables
DOCKER_IMAGE=ipquorum-agent
DOCKER_TAG=${VERSION}

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	@echo "Building ${BINARY_NAME}..."
	@go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/agent
	@echo "Build complete: bin/${BINARY_NAME}"

build-linux: ## Build for Linux
	@echo "Building ${BINARY_NAME} for Linux..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux ./cmd/agent
	@echo "Build complete: bin/${BINARY_NAME}-linux"

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -f ${BINARY_NAME}
	@echo "Clean complete"

run: build ## Build and run the agent
	@echo "Running ${BINARY_NAME}..."
	@./bin/${BINARY_NAME}

run-config: build ## Run with config file
	@echo "Running ${BINARY_NAME} with config..."
	@./bin/${BINARY_NAME} --config config.yaml

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build \
		--build-arg VERSION=${VERSION} \
		--build-arg BUILD_TIME=${BUILD_TIME} \
		--build-arg GIT_COMMIT=${GIT_COMMIT} \
		-t ${DOCKER_IMAGE}:${DOCKER_TAG} \
		-t ${DOCKER_IMAGE}:latest \
		.
	@echo "Docker image built: ${DOCKER_IMAGE}:${DOCKER_TAG}"

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run --rm -it \
		-p 8080:8080 \
		-p 9090:9090 \
		-v /var/run/docker.sock:/var/run/docker.sock \
		${DOCKER_IMAGE}:${DOCKER_TAG}

docker-run-podman: ## Run Docker container with Podman socket
	@echo "Running Docker container with Podman..."
	@docker run --rm -it \
		-p 8080:8080 \
		-p 9090:9090 \
		-v /run/podman/podman.sock:/run/podman/podman.sock \
		${DOCKER_IMAGE}:${DOCKER_TAG}

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete"

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...
	@echo "Vet complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded"

install: build ## Install binary to $GOPATH/bin
	@echo "Installing ${BINARY_NAME}..."
	@go install ${LDFLAGS} ./cmd/agent
	@echo "Install complete"

version: ## Show version information
	@echo "Version:    ${VERSION}"
	@echo "Build Time: ${BUILD_TIME}"
	@echo "Git Commit: ${GIT_COMMIT}"

all: clean deps fmt vet test build ## Run all checks and build

# Made with Bob
