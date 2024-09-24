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
# LDFLAGS=-ldflags="-s -w"
# BUILD_FLAGS=-mod=readonly $(LDFLAGS)

BUILD_FLAGS=-mod=readonly -ldflags="-s -w -X intercept/cmd.buildVersion=$(GIT_TAG)-$(GIT_COMMIT)" 

# All compilation platforms
ALL_PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm linux/arm64 windows/amd64

# Docker build platforms
DOCKER_PLATFORMS=linux/amd64 linux/arm linux/arm64

# Docker image name
DOCKER_IMAGE=test/intercept

# Main build target
all: clean build-all

# Clean
clean:
	$(GOCLEAN)
	rm -f release/$(BINARY_NAME)*

# Run tests
test:
	$(GOTEST) -v ./...

# Simple build for the current platform
build: clean
	CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o release/$(BINARY_NAME) .

# Compress binary using UPX
compress-binary:
	- docker run --rm -w $(shell pwd) -v $(shell pwd):$(shell pwd) docker.io/xfhg/upx:latest -9 release/$(BINARY_NAME)

#Checksums
sha256sums:
	@for file in release/*; do \
		echo "Generating SHA256 for $$file"; \
		sha256sum "$$file" > "$$file.sha256"; \
	done

# Build for all platforms
build-all: clean $(ALL_PLATFORMS) sha256sums

define BUILD_PLATFORM
build-$(1)-$(2):
	CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) $(GOBUILD) $(BUILD_FLAGS) -o release/$(BINARY_NAME)-$(1)-$(2)$(if $(filter windows,$(1)),.exe,) .
	$(MAKE) compress-binary BINARY_NAME=$(BINARY_NAME)-$(1)-$(2)$(if $(filter windows,$(1)),.exe,)

.PHONY: build-$(1)-$(2)
endef

$(foreach platform,$(ALL_PLATFORMS),$(eval $(call BUILD_PLATFORM,$(word 1,$(subst /, ,$(platform))),$(word 2,$(subst /, ,$(platform))))))

$(ALL_PLATFORMS): 
	$(MAKE) build-$(word 1,$(subst /, ,$@))-$(word 2,$(subst /, ,$@))

# Docker build commands for specific platforms
define DOCKER_BUILD_PLATFORM
docker-build-$(1)-$(2): build-$(1)-$(2)
	docker buildx build --platform $(1)/$(2) \
		--build-arg BINARY=release/$(BINARY_NAME)-$(1)-$(2) \
		-t $(DOCKER_IMAGE):$(1)-$(2) \
		--load \
		.

.PHONY: docker-build-$(1)-$(2)
endef

$(foreach platform,$(DOCKER_PLATFORMS),$(eval $(call DOCKER_BUILD_PLATFORM,$(word 1,$(subst /, ,$(platform))),$(word 2,$(subst /, ,$(platform))))))

# Build Docker images for all specified Linux platforms
docker-build-all: $(foreach platform,$(DOCKER_PLATFORMS),docker-build-$(word 1,$(subst /, ,$(platform)))-$(word 2,$(subst /, ,$(platform))))

.PHONY: all clean build build-all test docker-build-all compress-binary $(ALL_PLATFORMS) sha256sums
