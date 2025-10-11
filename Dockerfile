# Step 1: Build the Go binary
FROM golang:1.23 AS builder

WORKDIR /app
COPY . .
RUN go build -o app ./cmd/app/main.go

# Step 2: Run in minimal image
FROM alpine:latest

WORKDIR /root/
COPY --from=builder /app/app .

EXPOSE 3000
CMD ["./app"]
