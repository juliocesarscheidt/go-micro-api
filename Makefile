SHELL=/bin/bash
PACKAGES := $(shell cd src/ && go list ./)
# docker variables
DOCKER_BUILDKIT=1
BUILDKIT_PROGRESS=plain
# api variables
API_NAME?=go-micro-api
API_VERSION?=v1.0.0
API_MESSAGE?=Hello World
API_ENVIRONMENT?=production

all: help

.PHONY: help
help: Makefile
	@echo
	@echo " Choose a make command to run"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo

## go-vet: vet code with go CLI
.PHONY: go-vet
go-vet:
	cd src/; \
	go vet $(PACKAGES)

## go-test: run unit tests with go CLI
.PHONY: go-test
go-test:
	cd src/; \
	go test -race -cover $(PACKAGES)

## go-build: build a binary with go CLI
.PHONY: go-build
go-build:
	cd src/; \
	GOOS=linux GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o main .

## go-run: run the API with go CLI
.PHONY: go-run
go-run:
	cd src/; \
	MESSAGE="$(API_MESSAGE)" ENVIRONMENT="$(API_ENVIRONMENT)" go run main.go

## docker-build: build the docker image
.PHONY: docker-build
docker-build:
	docker image build --tag "juliocesarmidia/$(API_NAME):$(API_VERSION)" ./src

## docker-push: push the docker image
.PHONY: docker-push
docker-push:
	docker image push "juliocesarmidia/$(API_NAME):$(API_VERSION)"

## docker-run: run the docker container
.PHONY: docker-run
docker-run:
	docker container run -d \
		--name $(API_NAME) \
		--publish 9000:9000 \
		--cap-drop ALL \
		--memory='256MB' \
		--cpus='0.5' \
		--env MESSAGE="$(API_MESSAGE)" \
		--env ENVIRONMENT="$(API_ENVIRONMENT)" \
		--restart on-failure \
		"juliocesarmidia/$(API_NAME):$(API_VERSION)"

## docker-logs: get logs from docker container
.PHONY: docker-logs
docker-logs:
	docker container logs -f --tail 100 $(API_NAME)

## docker-exec: exec a command into docker container
.PHONY: docker-exec
docker-exec:
	docker container exec -it $(API_NAME) /bin/sh

## docker-rm: remove running docker container
.PHONY: docker-rm
docker-rm:
	docker container rm -f $(API_NAME)

## helm-install: install the helm release
.PHONY: helm-install
helm-install:
	helm upgrade -i "$(API_NAME)" ./helm --debug --wait --timeout 15m \
		--set configmaps.MESSAGE="$(API_MESSAGE)" \
		--set configmaps.ENVIRONMENT="$(API_ENVIRONMENT)"

## helm-uninstall: uninstall the helm release
.PHONY: helm-uninstall
helm-uninstall:
	helm delete "$(API_NAME)" --debug
