GO_CODE_PATH=./...

.DEFAULT_GOAL := explain
.PHONY: explain
explain:
	#### Travel API client
	#
	### Targets
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

##
# Setup targets
##

.PHONY: setup
setup: clean install build ## Set up for development

.PHONY: install
install: ## Install any specific tooling
ifeq ($(CI),true)
	npm ci
else
	npm install
endif
	go install github.com/golang/mock/mockgen@v1.5.0
	go generate $(GO_CODE_PATH)

.PHONY: clean
clean: ## Clean the local filesystem
	rm -fr node_modules
	git clean -fdX


##
## Vet targets
##

.PHONY: vet
vet: vet-go prettier ## Vet the code

.PHONY: vet-go
vet-go: ## Vet the Go code
	@echo "Vet the Go code..."
	go vet -v $(GO_CODE_PATH)

.PHONY: lint-go
lint-go: ## Lint the Go code
	@echo "Lint the Go code..."
	golangci-lint run -v

.PHONY: prettier
prettier: ## Run Prettier
	@echo "Run Prettier"
	npx prettier --check .


##
# Build targets
##

.PHONY: build
build: build-go ## Build everything

.PHONY: build-go
build-go: ## Build the Go code
	go build $(GO_CODE_PATH)


##
# Test targets
##

.PHONY: test
test: test-go ## Run all the tests

.PHONY: test-go
test-go: ## Run the Go tests
	go test $(GO_CODE_PATH) -coverprofile=coverage.out
	go tool cover -func=coverage.out
