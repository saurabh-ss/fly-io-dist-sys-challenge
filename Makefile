.PHONY: all build install clean test help

# Binary names
BINARIES := maelstrom-echo maelstrom-unique-ids maelstrom-broadcast

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

build: ## Build all binaries to bin/
	@echo "Building binaries..."
	@mkdir -p $(BUILD_DIR)
	@for binary in $(BINARIES); do \
		echo "  Building $$binary..."; \
		$(GOBUILD) -o $(BUILD_DIR)/$$binary ./cmd/$$binary || exit 1; \
	done
	@echo "✓ Build complete"

install: build ## Install binaries to $(INSTALL_DIR)
	@echo "Installing binaries to $(INSTALL_DIR)..."
	@mkdir -p $(INSTALL_DIR)
	@for binary in $(BINARIES); do \
		echo "  Installing $$binary..."; \
		cp $(BUILD_DIR)/$$binary $(INSTALL_DIR)/$$binary; \
		chmod +x $(INSTALL_DIR)/$$binary; \
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

# Individual binary targets
$(BUILD_DIR)/maelstrom-echo: ## Build maelstrom-echo
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $@ ./cmd/maelstrom-echo

$(BUILD_DIR)/maelstrom-unique-ids: ## Build maelstrom-unique-ids
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $@ ./cmd/maelstrom-unique-ids

$(BUILD_DIR)/maelstrom-broadcast: ## Build maelstrom-broadcast
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $@ ./cmd/maelstrom-broadcast

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

