# Makefile - Directives for build automation.
# Copyright (c) 2018 - 2021  Richard Huang <rickypc@users.noreply.github.com>
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at https://mozilla.org/MPL/2.0/.

.PHONY: clean
clean: ## Tidy module and remove temporary files
	@go mod tidy
	@go clean

.PHONY: cover
cover: test ## Run all the test and generate test coverage report
	go tool cover -html=cp.out

.PHONY: format
format: ## Run code formatter
	@gofmt -d -s .

.PHONY: format-auto
format-auto: ## Run code auto-formatter. It will attempt to overwrite the files.
	@gofmt -w .

.PHONY: help
help: ## List of all available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint: ## Run code linters
	@golangci-lint run

.PHONY: ready
ready: format lint test ## Prepare for publish
	@echo "\033[1;32mgit commit && git push origin && git tag vNEW-VERSION && git push --tags\033[0m"; \

.PHONY: test
test: clean version ## Run all the tests
	go test -coverprofile cp.out ./...

.PHONY: version
version: ## Display go version
	@go version

.DEFAULT_GOAL := help
