# enable BASH-specific features
SHELL := /bin/bash

SOURCE_DIR := $(shell pwd)

GOFILES!=find . -name '*.go'

.PHONY: build
build: cubit

.PHONY: test
test:
	@echo "Running tests..."
	@go test -race $$(go list ./...)

.PHONY: coverage
coverage:
	@echo "Running coverage..."
	@go test -race -coverprofile=coverage.txt -covermode=atomic $$(go list ./...)

.PHONY: vet
vet:
	@echo "Running vet..."
	@go vet $$(go list ./...)

.PHONY: lint
lint:
	@echo "Running lint..."
	@golint $$(go list ./...)

.PHONY: clean
clean:
	@echo "Running clean..."
	@go clean

go.sum: $(GOFILES) go.mod
	go mod tidy

cubit: $(GOFILES) go.mod go.sum
	@echo "Building binary..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-o $@ \
		-a \
		-ldflags '-extldflags "-static"'

.PHONY: init
init:
	@echo "Building docker volume..."
	@docker volume create mysql_database

.PHONY: prune
prune:
	@echo "Prune docker volume..."
	@docker volume rm mysql_database

.PHONY: env-up
env-up:
	@docker-compose -f docker-compose.yml up -d --force-recreate

.PHONY: env-down
env-down:
	@docker-compose -f docker-compose.yml down --remove-orphans

.PHONY: standalone-up
standalone-up:
	@docker-compose -f standalone.yml up -d --force-recreate

.PHONY: standalone-down
standalone-down:
	@docker-compose -f standalone.yml down --remove-orphans
