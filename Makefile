.PHONY: build run test mock vet lint clean docs help migrate-create migrate-up migrate-down check

BUILD_DIR := ./bin

## build: Compile the application binary
build:
	./scripts/build.sh $(BUILD_DIR)

## run: Build and run the application
run: build
	./scripts/run.sh $(BUILD_DIR)

## test: Run all tests
test:
	./scripts/test.sh

## mock: Generate test mocks
mock:
	./scripts/mock.sh

## vet: Run go vet
vet:
	./scripts/vet.sh

## lint: Run static analysis (requires golangci-lint)
lint:
	./scripts/lint.sh

## clean: Remove build artifacts
clean:
	./scripts/clean.sh $(BUILD_DIR)

## docs: Generate swagger documentation
docs:
	./scripts/docs.sh

## migrate-create: Create a new migration file (e.g., make migrate-create name=create_users_table)
migrate-create:
	./scripts/migrate.sh create "" $(name)

## migrate-up: Run migrations up (e.g., make migrate-up dsn="postgres://user:pass@host:port/db?sslmode=disable")
migrate-up:
	./scripts/migrate.sh up $(dsn) $(steps)

## migrate-down: Run migrations down (e.g., make migrate-down dsn="postgres://user:pass@host:port/db?sslmode=disable")
migrate-down:
	./scripts/migrate.sh down $(dsn) $(steps)

## help: Display available targets
help:
	./scripts/help.sh

## check: Run all checks (docs, vet, lint, test)
check: docs vet lint test
