.PHONY: all build clean test help

# Binary names
BINARIES := maelstrom-echo maelstrom-unique-ids maelstrom-broadcast maelstrom-counter maelstrom-kafka

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

build: $(addprefix $(BUILD_DIR)/,$(BINARIES)) ## Build all binaries to bin/ and install to $(INSTALL_DIR)
	@echo "✓ Build complete"

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
$(BUILD_DIR)/maelstrom-echo: $(wildcard cmd/maelstrom-echo/*.go) $(wildcard *.go) go.mod go.sum ## Build maelstrom-echo
	@echo "  Building maelstrom-echo..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $@ ./cmd/maelstrom-echo
	@mkdir -p $(INSTALL_DIR)
	@cp $@ $(INSTALL_DIR)/maelstrom-echo
	@chmod +x $(INSTALL_DIR)/maelstrom-echo
	@echo "  Installed maelstrom-echo to $(INSTALL_DIR)"

$(BUILD_DIR)/maelstrom-unique-ids: $(wildcard cmd/maelstrom-unique-ids/*.go) $(wildcard *.go) go.mod go.sum ## Build maelstrom-unique-ids
	@echo "  Building maelstrom-unique-ids..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $@ ./cmd/maelstrom-unique-ids
	@mkdir -p $(INSTALL_DIR)
	@cp $@ $(INSTALL_DIR)/maelstrom-unique-ids
	@chmod +x $(INSTALL_DIR)/maelstrom-unique-ids
	@echo "  Installed maelstrom-unique-ids to $(INSTALL_DIR)"

$(BUILD_DIR)/maelstrom-broadcast: $(wildcard cmd/maelstrom-broadcast/*.go) $(wildcard *.go) go.mod go.sum ## Build maelstrom-broadcast
	@echo "  Building maelstrom-broadcast..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $@ ./cmd/maelstrom-broadcast
	@mkdir -p $(INSTALL_DIR)
	@cp $@ $(INSTALL_DIR)/maelstrom-broadcast
	@chmod +x $(INSTALL_DIR)/maelstrom-broadcast
	@echo "  Installed maelstrom-broadcast to $(INSTALL_DIR)"

$(BUILD_DIR)/maelstrom-counter: $(wildcard cmd/maelstrom-counter/*.go) $(wildcard *.go) go.mod go.sum ## Build maelstrom-counter
	@echo "  Building maelstrom-counter..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $@ ./cmd/maelstrom-counter
	@mkdir -p $(INSTALL_DIR)
	@cp $@ $(INSTALL_DIR)/maelstrom-counter
	@chmod +x $(INSTALL_DIR)/maelstrom-counter
	@echo "  Installed maelstrom-counter to $(INSTALL_DIR)"

$(BUILD_DIR)/maelstrom-kafka: $(wildcard cmd/maelstrom-kafka/*.go) $(wildcard *.go) go.mod go.sum ## Build maelstrom-kafka
	@echo "  Building maelstrom-kafka..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $@ ./cmd/maelstrom-kafka
	@mkdir -p $(INSTALL_DIR)
	@cp $@ $(INSTALL_DIR)/maelstrom-kafka
	@chmod +x $(INSTALL_DIR)/maelstrom-kafka
	@echo "  Installed maelstrom-kafka to $(INSTALL_DIR)"

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

