FROM golang:1.21-alpine AS base

USER root

# Create app user
RUN useradd -m -s /bin/bash app

WORKDIR /api

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./cmd ./cmd
COPY ./pkg ./pkg

RUN mkdir tmp
RUN chown -R app tmp
RUN mkdir app
RUN chown -R app app
RUN chown -R app /api
ENV GOCACHE=/tmp/

USER app

FROM base AS builder-prod
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

FROM base AS test
RUN go test ./...

FROM alpine:3.18 AS app

USER root

RUN apk add ca-certificates --no-cache

# Create app user
RUN adduser -D -s /bin/sh app

WORKDIR /api
COPY --from=builder-prod /api/main .
RUN chmod +x main

USER app

ENTRYPOINT ["./main"]