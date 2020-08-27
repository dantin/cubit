# enable BASH-specific features
SHELL := /bin/bash

SOURCE_DIR := $(shell pwd)

GOFILES!=find . -name '*.go'
GOLDFLAGS := -s -w -extldflags $(LDFLAGS)

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
	@go build \
		-trimpath \
		-o $@ \
		-ldflags "$(GOLDFLAGS)"

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
