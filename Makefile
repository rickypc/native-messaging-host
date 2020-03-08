.PHONY: clean
clean: ## Remove temporary files
	@go mod tidy
	@go clean

.PHONY: cover
cover: test ## Run all the test and generate test coverage report
	go tool cover -html=cp.out

.PHONY: format
format: ## Run code formatter
	@gofmt -d .

.PHONY: help
help: ## List of all available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint: ## Run code linters
	@golangci-lint run

.PHONY: test
test: clean version ## Run all the tests
	go test -coverprofile cp.out

.PHONY: version
version: ## Display go version
	@go version

.DEFAULT_GOAL := help
