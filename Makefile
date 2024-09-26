# Binary name
BINARY_NAME=intercept

# Binary Version
GIT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.X")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

# Build flags
BUILD_FLAGS=-mod=readonly -ldflags="-s -w -X intercept/cmd.buildVersion=$(GIT_TAG)-$(GIT_COMMIT)"

# All compilation platforms
PLATFORMS := \
    darwin/amd64 \
    darwin/arm64 \
    windows/amd64 \
    linux/amd64 \
    linux/arm64 \
    linux/arm/v7

# Docker build platforms
DOCKER_PLATFORMS := linux/amd64 linux/arm64 

# Docker image name
DOCKER_IMAGE=test/intercept

# Main build target
all: clean build-all

# Clean
clean:
	$(GOCLEAN)
	rm -rf release/$(BINARY_NAME)*

# Run tests
test:
	$(GOTEST) -v ./...

# Simple build for the current platform
build: clean
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o release/$(BINARY_NAME) .

# Compress binary using UPX
compress-binary:
	- docker run --rm -v $(shell pwd):/workspace -w /workspace docker.io/xfhg/upx:latest -9 release/$(BINARY_NAME)

# Checksums
sha256sums:
	@for file in release/$(BINARY_NAME)*; do \
		echo "Generating SHA256 for $$file"; \
		sha256sum "$$file" > "$$file.sha256"; \
	done

# Build for all platforms
build-all: clean $(PLATFORMS) sha256sums

define BUILD_PLATFORM
build-$(1):
	@echo "Building for platform $(1)"
	$(eval GOOS := $(word 1, $(subst /, ,$(1))))
	$(eval GOARCH := $(word 2, $(subst /, ,$(1))))
	$(eval GOARM := $(word 3, $(subst /, ,$(1))))
	$(eval BIN_SUFFIX := $(if $(GOARM),-$(GOARM),))
	$(eval OUTPUT_NAME := $(BINARY_NAME)-$(GOOS)-$(GOARCH)$(BIN_SUFFIX)$(if $(filter windows,$(GOOS)),.exe,))
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(if $(GOARM),GOARM=$(GOARM),) \
		$(GOBUILD) $(BUILD_FLAGS) -o release/$(OUTPUT_NAME) .
	$(MAKE) compress-binary BINARY_NAME=$(OUTPUT_NAME)
.PHONY: build-$(1)
endef

$(foreach platform,$(PLATFORMS),$(eval $(call BUILD_PLATFORM,$(platform))))

$(PLATFORMS):
	$(MAKE) build-$@

# Docker build commands for specific platforms
define DOCKER_BUILD_PLATFORM
docker-build-$(1):
	$(MAKE) build-$(1)
	$(eval GOOS := $(word 1, $(subst /, ,$(1))))
	$(eval GOARCH := $(word 2, $(subst /, ,$(1))))
	$(eval GOARM := $(word 3, $(subst /, ,$(1))))
	$(eval BIN_SUFFIX := $(if $(GOARM),-v$(GOARM),))
	$(eval OUTPUT_NAME := $(BINARY_NAME)-$(GOOS)-$(GOARCH)$(BIN_SUFFIX)$(if $(filter windows,$(GOOS)),.exe,))
	docker buildx build --platform $(GOOS)/$(GOARCH)$(if $(GOARM),/v$(GOARM),) \
		--build-arg BINARY=release/$(OUTPUT_NAME) \
		-t $(DOCKER_IMAGE):$(GOOS)-$(GOARCH)$(BIN_SUFFIX) \
		--load \
		.
.PHONY: docker-build-$(1)
endef

$(foreach platform,$(DOCKER_PLATFORMS),$(eval $(call DOCKER_BUILD_PLATFORM,$(platform))))

# Build Docker images for all specified platforms
docker-build-all: $(foreach platform,$(DOCKER_PLATFORMS),docker-build-$(platform))

.PHONY: all clean build build-all compress-binary docker-build-all $(PLATFORMS) sha256sums