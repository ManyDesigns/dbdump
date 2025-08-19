# Makefile for the Go Database Backup Tool

# Define the name of the final binary
BINARY_NAME=dbdump

# Define the path to the main package
SRC_DIR=./cmd/dbdump

# Go linker flags to reduce binary size for release builds.
# -s: Omit the symbol table.
# -w: Omit the DWARF debug information.
LDFLAGS_RELEASE = -ldflags="-s -w"

# Default target executed when you just run `make`
.PHONY: all
all: build

# Build the application binary for development (includes debug info)
.PHONY: build
build: tidy
	@echo "Building $(BINARY_NAME) for development..."
	@go build -o $(BINARY_NAME) $(SRC_DIR)
	@echo "$(BINARY_NAME) built successfully."

# Build a small, optimized release binary
# This strips debug information to significantly reduce the file size.
.PHONY: build-release
build-release: tidy
	@echo "Building $(BINARY_NAME) for release (optimized for size)..."
	@go build $(LDFLAGS_RELEASE) -o $(BINARY_NAME) $(SRC_DIR)
	@echo "Release binary built successfully."

# Run the application
# Example: make run ARGS="--all --type=postgres --pgpassword=your_pass"
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME) with arguments: $(ARGS)"
	@./$(BINARY_NAME) $(ARGS)

# Tidy Go modules
.PHONY: tidy
tidy:
	@echo "Tidying Go modules..."
	@go mod tidy

# Clean up build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -f $(BINARY_NAME)
	@echo "Cleanup complete."

# Display help message
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build         - Build the application binary for development."
	@echo "  build-release - Build a small, optimized release binary."
	@echo "  run           - Build and run the application. Use ARGS=\"...\" to pass flags."
	@echo "                Example: make run ARGS=\"--all --pgpassword=secret\""
	@echo "  tidy          - Synchronize go.mod dependencies."
	@echo "  clean         - Remove the compiled binary."
	@echo "  help          - Display this help message."

