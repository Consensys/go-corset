GOLANGCI_VERSION:=1.57.1
PROJECT_NAME:=go-corset
GOPATH_BIN:=$(shell go env GOPATH)/bin

.PHONY: install
install:
	# Install golangci-lint for go code linting.
	curl -sSfL \
		"https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" | \
		sh -s -- -b ${GOPATH_BIN} v${GOLANGCI_VERSION}

.PHONY: all
all: clean lint test build

.PHONY: lint
lint:
	@echo ">>> Performing golang code linting.."
	golangci-lint run --config=.golangci.yml

.PHONY: test
test:
	@echo ">>> Running Unit Tests..."
	go test -v -race ./...

.PHONY: test-no-cache
test-no-cache:
	@echo ">>> Running Unit Tests without Caching..."
	go test -v -count=1 -race ./...

.PHONY: cover-test
cover-test:
	@echo ">>> Running Tests with Coverage..."
	go test -v -race ./... -coverprofile=coverage.out -covermode=atomic

.PHONY: show-cover
show-cover:
	@go tool cover -html=coverage.out

.PHONY: build
build:
	@echo ">>> Building ${PROJECT_NAME} API server..."
	go build -o bin/${PROJECT_NAME} cmd/${PROJECT_NAME}/main.go

.PHONY: clean
clean:
	@echo ">>> Removing old binaries and env files..."
	@rm -rf bin/*
	@rm -rf .env
