# syntax=docker/dockerfile:1

# Multi-stage Dockerfile for Go Echo app with Clean Architecture
# Stages:
# - builder: build static binary
# - runtime: minimal image to run binary (alpine)
# - dev: development image with Air hot-reload

ARG GO_VERSION=1.22

FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Build static binary for Linux amd64
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/app

FROM alpine:3.20 AS runtime
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 appuser
COPY --from=builder /out/app /app/app
EXPOSE 3000
ENV APP_ENV=production
USER appuser
ENTRYPOINT ["/app/app"]

# Development stage with Air hot reload
FROM golang:${GO_VERSION}-alpine AS dev
WORKDIR /src
RUN apk add --no-cache git bash ca-certificates tzdata \
    && go install github.com/cosmtrek/air@latest
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["air", "-c", ".air.toml"]