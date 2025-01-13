# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=relayer
BUILD_DIR=build

# Build the binary
build:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/main.go

# Clean the binary
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run the binary
run: build
	./$(BUILD_DIR)/$(BINARY_NAME) 
# Default target
all: build

.PHONY: build clean run all

# Debug build for debugging with dlv or any other debugger
debug:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/main.go


# Install the binary
install:
	@echo "Installing $(BINARY_NAME)..."
	@if [ "$$(uname)" = "Darwin" ]; then \
		$(GOBUILD) -o /usr/local/bin/$(BINARY_NAME) -v ./cmd/main.go; \
	elif [ "$$(uname)" = "Linux" ]; then \
		$(GOBUILD) -o /usr/local/bin/$(BINARY_NAME) -v ./cmd/main.go; \
	else \
		echo "Unsupported operating system. Please install manually."; \
		exit 1; \
	fi
	@echo "$(BINARY_NAME) has been installed to /usr/local/bin/$(BINARY_NAME)"

.PHONY: install


###############################################################################
###                                  Tests                                  ###
###############################################################################

test: test-unit
test-all: test-unit test-race
PACKAGES_UNIT=$(shell go list ./... | grep -Ev 'vendor|importer')
TEST_PACKAGES=./...
TEST_TARGETS := test-unit test-unit-cover test-race

# Test runs-specific rules. To add a new test target, just add
# a new rule, customise ARGS or TEST_PACKAGES ad libitum, and
# append the new rule to the TEST_TARGETS list.
test-unit: ARGS=-timeout=10m -race
test-unit: TEST_PACKAGES=$(PACKAGES_UNIT)

test-race: ARGS=-race
test-race: TEST_PACKAGES=$(PACKAGES_NOSIMULATION)
$(TEST_TARGETS): run-tests

test-unit-cover: ARGS=-timeout=10m -race -coverprofile=coverage.txt -covermode=atomic
test-unit-cover: TEST_PACKAGES=$(PACKAGES_UNIT)

run-tests:
ifneq (,$(shell which tparse 2>/dev/null))
	go test -mod=readonly -json $(ARGS) $(EXTRA_ARGS) $(TEST_PACKAGES) | tparse
else
	go test -mod=readonly $(ARGS)  $(EXTRA_ARGS) $(TEST_PACKAGES)
endif

.PHONY: run-tests test test-all $(TEST_TARGETS)

# Implements test splitting and running. This is pulled directly from
# the github action workflows for better local reproducibility.

GO_TEST_FILES != find $(CURDIR) -name "*_test.go"

# default to four splits by default
NUM_SPLIT ?= 3

UNAME_S := $(shell uname -s)

ifeq ($(UNAME_S),Darwin)  # macOS
    SPLIT_CMD = gsplit
else
    SPLIT_CMD = split
endif

BUILDDIR ?= $(CURDIR)/build

$(BUILDDIR):
	mkdir -p $@

# The format statement filters out all packages that don't have tests.
# Note we need to check for both in-package tests (.TestGoFiles) and
# out-of-package tests (.XTestGoFiles).
$(BUILDDIR)/packages.txt:$(GO_TEST_FILES) $(BUILDDIR)
	go list -f "{{ if (or .TestGoFiles .XTestGoFiles) }}{{ .ImportPath }}{{ end }}" ./... | sort > $@

split-test-packages:$(BUILDDIR)/packages.txt
	$(SPLIT_CMD) -d -n l/$(NUM_SPLIT) $< $<.

test-group-%:split-test-packages
	cat $(BUILDDIR)/packages.txt.$* | xargs go test -mod=readonly -timeout=10m -race -coverprofile=$(BUILDDIR)/$*.profile.out
