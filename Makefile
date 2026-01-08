.PHONY: setup
setup:
	@go mod tidy

.PHONY: gen
gen:
	@goctl api go --api ./dsl/auth.api --dir ./ --style goZero
	@goctl model mysql ddl --src ./model/mysql/user.sql --dir ./model/mysql

.PHONY: build
build:
	@go build -ldflags="-s -w" -o ./bin/auth-api ./auth.go

.PHONY: run
run:
	@go run auth.go -f etc/auth-api.yaml

.PHONY: test
test:
	@echo "Running tests..."
	@go test ./tests/... -v -cover

.PHONY: coverage
coverage:
	@echo "Running tests with coverage..."
	@go test ./tests/... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: clean
clean:
	@rm -f coverage.out coverage.html
	@go clean -cache
	@echo "Cleaned test artifacts"
