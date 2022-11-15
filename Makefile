PROJECT_NAME := eci-profile
all: build

PHONY: build
build:
	env GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o bin/${PROJECT_NAME} ./cmd

PHONY: clean
clean:
	rm -f ./bin/${PROJECT_NAME}

SOURCE_DIRS = cmd pkg
SOURCE_PACKAGES = ./cmd/... ./pkg/...
TEST_FILES := ./...

PHONY: fmt
fmt: ## Run go fmt and modify files in place
	@gofmt -s -w $(SOURCE_DIRS)

.PHONY: gofmt
gofmt: ## Run go fmt and list the files differs from gofmt's
	@gofmt -s -l $(SOURCE_DIRS)
	@test -z "`gofmt -s -l $(SOURCE_DIRS)`"

.PHONY: vet
vet: ## Run go vet
	@go vet $(SOURCE_PACKAGES)

.PHONY: test
test: $(GOTEST_FILES)
	@go test $(TEST_FILES)

.PHONY: imports
imports: ## Run goimports and modify files in place
	@goimports -w $(SOURCE_DIRS)

.PHONY: goimports
goimports: ## Run goimports and list the files differs from goimport's
	@goimports -l $(SOURCE_DIRS)
	@test -z "`goimports -l $(SOURCE_DIRS)`"

.PHONY: golint
golint: ## Run golint
	@golint -set_exit_status $(SOURCE_PACKAGES)

.PHONY: gocyclo
gocyclo: ## Run gocyclo (calculates cyclomatic complexities)
	@gocyclo -over 15 `find $(SOURCE_DIRS) -type f -name "*.go"`