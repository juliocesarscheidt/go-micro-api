FROM golang:1.18-alpine as builder
LABEL maintainer="Julio Cesar <julio@blackdevs.com.br>"

WORKDIR /go/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./

RUN go install golang.org/x/lint/golint@latest
RUN ls -lth /go/bin/
RUN echo $GOPATH
RUN ls -lth $GOPATH/bin/

EXPOSE 9000
# set default env
ENV ENVIRONMENT=development

ENTRYPOINT []
CMD ["go", "run", "main.go"]
