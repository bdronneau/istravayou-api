SHELL = /bin/sh

APP_NAME ?= istravayou-api
PACKAGES ?= ./...
GO_FILES ?= */*/*.go
@VERSION = $(shell git describe --tags)

GOBIN=bin
BINARY_PATH=$(GOBIN)/$(APP_NAME)

LIB_SOURCE = cmd/istravayou/main.go

GO_ARCH ?= $(shell go env GOHOSTARCH)
GO_OS ?= $(shell go env GOHOSTOS)

.DEFAULT_GOAL := app

## help: Display list of commands
.PHONY: help
help: Makefile
	@sed -n 's|^##||p' $< | column -t -s ':' | sed -e 's|^| |'

## name: Output name
.PHONY: name
name:
	@echo -n $(APP_NAME)

## version: Output sha1 of last commit
.PHONY: version
version:
	@echo -n $(VERSION)

##app: Build app with dependencies download
.PHONY: app
app: deps go

## go: Build app
.PHONY: go
go: format lint build

## deps: Download dependencies
.PHONY: deps
deps:
	go get github.com/kisielk/errcheck
	go get golang.org/x/lint/golint
	go get golang.org/x/tools/cmd/goimports

## format: Format code
.PHONY: format
format:
	goimports -w $(GO_FILES)
	gofmt -s -w $(GO_FILES)

## lint: Lint code
.PHONY: lint
lint:
	golint $(PACKAGES)
	errcheck -verbose -ignoretests $(PACKAGES)
	go vet $(PACKAGES)

## build: Build binary
.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix nocgo -o $(BINARY_PATH) $(LIB_SOURCE)

# clean: Delete binary
.PHONY: clean
clean:
	rm $(BINARY_PATH)_$(VERSION)

.PHONY: all
all: build install
