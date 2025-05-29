# Makefile for GoSight builds

#---------------------------------------
# Configuration
#---------------------------------------
# Semantic version (override on CLI: make VERSION=0.1.0-alpha.1)
VERSION ?= dev

# Build metadata
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD)

# ldflags for injecting version info
LDFLAGS := -X 'main.Version=$(VERSION)' \
           -X 'main.BuildTime=$(BUILD_TIME)' \
           -X 'main.GitCommit=$(GIT_COMMIT)'

# Output directories
BIN_DIR := bin
SERVER_OUT := $(BIN_DIR)/gosight-server


#---------------------------------------
# Phony targets
#---------------------------------------
.PHONY: all server fmt test clean

# Default target builds both binaries
all: server


#---------------------------------------
# Run
#---------------------------------------
.PHONY: run
run: server
	@echo "Running GoSight server $(VERSION)"
	sudo  $(SERVER_OUT)

# Build the GoSight server
server:
	@mkdir -p $(BIN_DIR)
	@echo "Building server $(VERSION)"
	go build \
		-ldflags "$(LDFLAGS)" \
		-o $(SERVER_OUT) \
		./cmd/

# Format code
fmt:
	go fmt ./...

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf $(BIN_DIR)