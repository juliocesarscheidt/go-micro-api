SHELL=/bin/bash
PACKAGES := $(shell go list ./)
# docker variables
DOCKER_REPO?=docker.io/juliocesarmidia/http-simple-api
DOCKER_TAG?=v1.0.0
DOCKER_BUILDKIT=1
BUILDKIT_PROGRESS=plain
# application variables
MESSAGE?="Hello World"

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## vet: vet code
.PHONY: vet
vet:
	go vet $(PACKAGES)

## test: run unit tests
.PHONY: test
test:
	go test -race -cover $(PACKAGES)

## build: build a binary with go CLI
.PHONY: build
build:
	GOOS=linux GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o ./main .

## run: run the API
.PHONY: run
run:
	go run main.go

## docker-build: build the docker image
.PHONY: docker-build
docker-build:
	docker image build --tag $(DOCKER_REPO):$(DOCKER_TAG) .

## docker-push: push the docker image
.PHONY: docker-push
docker-push:
	docker image push $(DOCKER_REPO):$(DOCKER_TAG)
