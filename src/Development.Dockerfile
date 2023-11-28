FROM golang:1.18-alpine as builder
LABEL maintainer="Julio Cesar <julio@blackdevs.com.br>"

WORKDIR /go/src/app

RUN go install golang.org/x/lint/golint@latest
RUN go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

EXPOSE 9000
# set default env
ENV ENVIRONMENT=development

ENTRYPOINT []
# CMD ["go", "run", "main.go"]
CMD ["air", "-c", ".air.toml"]
