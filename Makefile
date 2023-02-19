SHELL=/bin/bash
PACKAGES := $(shell go list ./)
# docker variables
DOCKER_REPO?=docker.io/juliocesarmidia/go-micro-api
DOCKER_TAG?=v1.0.0
DOCKER_BUILDKIT=1
BUILDKIT_PROGRESS=plain
# application variables
MESSAGE?="Hello World"
# kubernetes variables
RELEASE_NAME?="go-micro-api"

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## vet: vet code with go CLI
.PHONY: vet
vet:
	go vet $(PACKAGES)

## test: run unit tests with go CLI
.PHONY: test
test:
	go test -race -cover $(PACKAGES)

## build: build a binary with go CLI
.PHONY: build
build:
	GOOS=linux GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o ./main .

## run: run the API with go CLI
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

## helm-install: install the helm release
.PHONY: helm-install
helm-install:
	helm upgrade -i "$(RELEASE_NAME)" ./helm --debug --wait --timeout 15m

## helm-uninstall: uninstall the helm release
.PHONY: helm-uninstall
helm-uninstall:
	helm delete "$(RELEASE_NAME)" --debug
