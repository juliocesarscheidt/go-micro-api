export DOCKER_REPO ?= juliocesarmidia/http-simple-api
export DOCKER_TAG ?= v1.0.0

export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

export BUILDKIT_PROGRESS=plain

SHELL=/bin/bash

all: vet fmt build

vet: main.go
	go vet . && echo $?

fmt: main.go
	go fmt . && echo $?

run: main.go
	go run main.go

build:
	docker image build --tag ${DOCKER_REPO}:${DOCKER_TAG} -f Dockerfile .

publish:
	docker image push ${DOCKER_REPO}:${DOCKER_TAG}
