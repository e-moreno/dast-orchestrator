FROM golang:1.23-alpine AS base

USER root

# Update packages and add security updates
RUN apk update && apk upgrade && apk add --no-cache ca-certificates

# Create app user
RUN adduser -D -s /bin/sh app

WORKDIR /api

COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy source code to match module structure
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
RUN go build -o main ./cmd/main.go

FROM base AS test
RUN go test ./...

FROM alpine:3.20 AS app

USER root

# Update packages and install ca-certificates
RUN apk update && apk upgrade && apk add --no-cache ca-certificates

# Create app user
RUN adduser -D -s /bin/sh app

USER app

WORKDIR /api
COPY --from=builder-prod /api/main .

ENTRYPOINT ["./main"]