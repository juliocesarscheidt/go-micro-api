FROM golang:1.18-alpine as builder
LABEL maintainer="Julio Cesar <julio@blackdevs.com.br>"

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN GOOS=linux GOARCH=amd64 GO111MODULE=on CGO_ENABLED=0 \
    go build -ldflags="-s -w" -o ./main .

FROM gcr.io/distroless/static:nonroot

LABEL maintainer="Julio Cesar <julio@blackdevs.com.br>"
LABEL org.opencontainers.image.source "https://github.com/juliocesarscheidt/go-micro-api"
LABEL org.opencontainers.image.description "Simple Golang API"
LABEL org.opencontainers.image.licenses "MIT"

WORKDIR /
COPY --from=builder /go/src/app/main .
USER nonroot:nonroot

EXPOSE 9000

ENTRYPOINT [ "/main" ]
