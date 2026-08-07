.PHONY: all build test test-coverage lint compose-up compose-down clean

GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
BINARY_DIR=bin

all: build test

build:
	@echo "Building binary..."
	mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/ ./...

test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

test-coverage:
	@echo "Running test coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

compose-up:
	@echo "Starting Docker Compose services..."
	docker-compose up -d

compose-down:
	@echo "Stopping Docker Compose services..."
	docker-compose down -v

clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR) coverage.out coverage.html
