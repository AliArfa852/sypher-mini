.PHONY: build run test clean rebuild extensions build-go docker docker-run docker-down

BINARY_NAME=sypher
BUILD_DIR=build
EXTENSIONS_DIR=extensions

# Default target
all: build

## extensions: Install and build all Node extensions (npm install + npm run build)
extensions:
	@for dir in $(EXTENSIONS_DIR)/*/; do \
		if [ -f "$$dir/package.json" ]; then \
			echo "Building extension: $$dir"; \
			(cd "$$dir" && npm install && npm run build) || exit 1; \
		fi; \
	done
	@echo "Extensions build complete"

## build: Build extensions and sypher binary
build: extensions
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/sypher
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## build-go: Build sypher binary only (skip extensions; use when extensions already built)
build-go:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/sypher
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## run: Build and run gateway
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) gateway

## test: Run tests
test:
	@go test ./...

## clean: Remove build artifacts, Go cache, and extension node_modules/dist
clean:
	@rm -rf $(BUILD_DIR)
	@go clean -cache
	@for dir in $(EXTENSIONS_DIR)/*/; do \
		if [ -f "$$dir/package.json" ]; then \
			rm -rf "$$dir/node_modules" "$$dir/dist"; \
		fi; \
	done
	@echo "Clean complete"

## rebuild: Clean then build
rebuild: clean build

## docker: Build Docker image
docker:
	@docker build -t sypher-mini:latest .
	@echo "Docker image built: sypher-mini:latest"

## docker-run: Run with docker-compose
docker-run:
	@docker-compose up -d
	@echo "Sypher-mini running. Health: http://localhost:18790/health"

## docker-down: Stop docker-compose
docker-down:
	@docker-compose down
