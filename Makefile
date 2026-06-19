.PHONY: build run test mock vet lint clean docs help migrate-create migrate-up migrate-down check githooks

## githooks: Configure git to use local githooks directory
githooks:
	git config core.hooksPath githooks
	chmod +x githooks/pre-commit githooks/pre-push

BUILD_DIR := ./bin

# Default config path
config ?= ./configs
CONFIG_ARG = -config $(config)

## build: Compile the application binary
build:
	go build -mod=mod -o $(BUILD_DIR)/main ./cmd/main.go
	@echo "Build completed. Binary is located at $(BUILD_DIR)/main"

## run: Build and run the application (e.g., make run config="./configs" or make run args="-config ./configs")
run: build
	$(BUILD_DIR)/main $(CONFIG_ARG) $(args)

## test: Run all tests
test:
	go test -v -race -count=1 ./...

## mock: Generate test mocks
mock:
	@echo "Mock Generation"
	@mkdir -p internal/core/usecase/mocks
	@rm -f internal/core/usecase/mocks/mock_*.go
	@for file in internal/core/ports/*.go; do \
		filename=$$(basename $$file); \
		mockname="mock_$$filename"; \
		echo "Generating mock for $$filename -> $$mockname"; \
		go run go.uber.org/mock/mockgen -source=$$file -destination=internal/core/usecase/mocks/$$mockname -package=mocks; \
	done
	@echo "Mocks generation completed successfully."

## vet: Run go vet
vet:
	go vet ./...

## lint: Run static analysis (requires golangci-lint)
lint:
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Error: 'golangci-lint' command not found." >&2; exit 1; }
	golangci-lint run -c .golangci.yml ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## docs: Generate swagger documentation
docs:
	@command -v swag >/dev/null 2>&1 || { echo "Error: 'swag' command not found." >&2; exit 1; }
	swag fmt
	swag init -g cmd/main.go -o api/swagger

## migrate-create: Create a new migration file (e.g., make migrate-create name=create_users_table)
migrate-create:
	@command -v migrate >/dev/null 2>&1 || { echo "Error: 'migrate' command not found." >&2; exit 1; }
	@if [ -z "$(name)" ]; then echo "Error: Name is required. Example: make migrate-create name=create_users_table" >&2; exit 1; fi
	migrate create -ext sql -dir db/migrations -seq $(name)

## migrate-up: Run migrations up (e.g., make migrate-up dsn="postgres://user:pass@host:port/db?sslmode=disable")
migrate-up:
	@command -v migrate >/dev/null 2>&1 || { echo "Error: 'migrate' command not found." >&2; exit 1; }
	@if [ -z "$(dsn)" ]; then echo "Error: DSN is required. Example: make migrate-up dsn=..." >&2; exit 1; fi
	migrate -path db/migrations -database "$(dsn)" up $(steps)

## migrate-down: Run migrations down (e.g., make migrate-down dsn="postgres://user:pass@host:port/db?sslmode=disable")
migrate-down:
	@command -v migrate >/dev/null 2>&1 || { echo "Error: 'migrate' command not found." >&2; exit 1; }
	@if [ -z "$(dsn)" ]; then echo "Error: DSN is required. Example: make migrate-down dsn=..." >&2; exit 1; fi
	migrate -path db/migrations -database "$(dsn)" down $(steps)

## help: Display available targets
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'

## check: Run all checks (docs, vet, lint, test)
check: docs vet lint test
