# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=loadbalancer
DOCKER_COMPOSE=docker-compose

.PHONY: all build clean test coverage deps docker-build docker-run docker-stop lint help

all: test build

build: ## Build the load balancer
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/loadbalancer

clean: ## Clean build files
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf test/reports/*

test: ## Run unit tests
	./scripts/run-tests.sh

coverage: ## Run tests with coverage
	$(GOTEST) -coverprofile=test/reports/coverage.out ./...
	$(GOCMD) tool cover -html=test/reports/coverage.out -o test/reports/coverage.html

deps: ## Download required dependencies
	$(GOMOD) download
	$(GOMOD) tidy

docker-build: ## Build docker images
	$(DOCKER_COMPOSE) build

docker-run: ## Run the application in docker
	$(DOCKER_COMPOSE) up

docker-stop: ## Stop docker containers
	$(DOCKER_COMPOSE) down

lint: ## Run linters
	golangci-lint run

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Initialize development environment
init: deps ## Initialize development environment
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vakenbolt/go-test-report@latest

# Run the load balancer locally
run: build ## Run the load balancer locally
	./$(BINARY_NAME)

# Create test reports directory
test-init: ## Create test reports directory
	mkdir -p test/reports

# Run integration tests
integration-test: ## Run integration tests
	$(GOTEST) -tags=integration ./... -v

# Generate all documentation
docs: ## Generate documentation
	$(GOCMD) doc -all > docs/API.md

# Default target
.DEFAULT_GOAL := help
