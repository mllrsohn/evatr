.DEFAULT_GOAL := help

.PHONY: test
test: ## Run all the tests
	go test -v -race -timeout=30s ./...

.PHONY: cover
cover: ## Run all the tests and opens the coverage report
	go test -covermode=atomic -coverprofile=coverage.txt -v -race  ./...
	go tool cover -html=coverage.txt

.PHONY: lint
lint: ## Runs the linter
	golangci-lint run --timeout 5m

# Help
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


.DEFAULT_GOAL := help
