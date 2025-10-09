# syntax=docker/dockerfile:1

# ARG default Go version
ARG GO_VERSION=1.25

# --------------------------
# Stage 1: Builder
# --------------------------
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

# Install dependencies for build
RUN apk add --no-cache ca-certificates git bash

# Copy Go modules manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary for target platform
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/app

# --------------------------
# Stage 2: Runtime
# --------------------------
FROM alpine:3.20 AS runtime

WORKDIR /app

# Install required packages and create a non-root user
RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 10001 appuser

# Copy binary from builder
COPY --from=builder /out/app /app/app

# Expose port
EXPOSE 3000

# Set environment for production
ENV APP_ENV=production

# Run as non-root
USER appuser

# Entry point
ENTRYPOINT ["/app/app"]

# --------------------------
# Stage 3: Development
# --------------------------
FROM golang:${GO_VERSION}-alpine AS dev
WORKDIR /src

RUN apk add --no-cache git bash ca-certificates tzdata \
    && go install github.com/cosmtrek/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]