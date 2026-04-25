# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \
    -o nanobot \
    ./cmd/nanobot/

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -u 1000 nanobot

WORKDIR /app

COPY --from=builder /build/nanobot .
COPY --from=builder /build/config ./config/
COPY --from=builder /build/storage ./storage/

# Create directories with correct permissions
RUN mkdir -p /app/storage/context /app/storage/logs /app/storage/memory \
    && chown -R nanobot:nanobot /app

USER nanobot

ENTRYPOINT ["./nanobot"]
