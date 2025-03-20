GOCORSET_VERSION:=$(shell git describe --always --tags)
GOCORSET_VERSION_PATH:="github.com/consensys/go-corset/pkg/cmd"
GOLANGCI_VERSION:=1.64.8
PROJECT_NAME:=go-corset
GOPATH_BIN:=$(shell go env GOPATH)/bin

.PHONY: install
install:
	# Install golangci-lint for go code linting.
	curl -sSfL \
		"https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" | \
		sh -s -- -b ${GOPATH_BIN} v${GOLANGCI_VERSION}
	# Install cobra-cli command generator.
	go install github.com/spf13/cobra-cli@latest

.PHONY: all
all: clean lint test build

.PHONY: lint
lint:
	@echo ">>> Performing golang code linting.."
	golangci-lint run --config=.golangci.yml

.PHONY: test
test:
	@echo ">>> Running Unit Tests..."
	go test --timeout 0 -v ./...

.PHONY: qtest
qtest:
	@echo ">>> Running (Quick) Tests..."
	go test -v -race --run Test_ ./...

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
	@echo ">>> Building ${PROJECT_NAME}... ${GOCORSET_VERSION}"
	go build -ldflags="-X '${GOCORSET_VERSION_PATH}.Version=${GOCORSET_VERSION}'" -o bin/${PROJECT_NAME} cmd/${PROJECT_NAME}/main.go

.PHONY: clean
clean:
	@echo ">>> Removing old binaries and env files..."
	@rm -rf bin/*
	@rm -rf .env
