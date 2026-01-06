.PHONY: all build install clean test tidy help

# Binary names
BINARIES := maelstrom-echo maelstrom-unique-ids maelstrom-broadcast maelstrom-counter maelstrom-kafka maelstrom-txn

# Build directory
BUILD_DIR := bin

# Install directory (defaults to GOPATH/bin, or /usr/local/bin if GOPATH not set)
INSTALL_DIR := $(or $(GOPATH),$(HOME)/go)/bin

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

all: build ## Build all binaries

build: $(addprefix $(BUILD_DIR)/,$(BINARIES)) ## Build all binaries to bin/
	@echo "✓ Build complete"

install: build ## Install all binaries to $(INSTALL_DIR)
	@echo "Installing binaries..."
	@mkdir -p $(INSTALL_DIR)
	@for binary in $(BINARIES); do \
		cp $(BUILD_DIR)/$$binary $(INSTALL_DIR)/$$binary; \
		chmod +x $(INSTALL_DIR)/$$binary; \
		echo "  Installed $$binary to $(INSTALL_DIR)"; \
	done
	@echo "✓ Install complete"

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@$(GOCLEAN)
	@echo "✓ Clean complete"

test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v ./...

tidy: ## Tidy go modules
	@echo "Tidying modules..."
	@$(GOMOD) tidy
	@echo "✓ Tidy complete"

# Pattern rule for building binaries
$(BUILD_DIR)/%: cmd/%/main.go go.mod go.sum
	@echo "  Building $*..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $@ ./cmd/$*

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
