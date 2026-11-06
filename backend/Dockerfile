# Stage 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
COPY go_sdk/ ./go_sdk/
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Install curl & tar to fetch pre-compiled migrate binary (much faster than compiling from source)
RUN apk add --no-cache git curl tar && \
    ARCH=$(apk --print-arch) && \
    if [ "$ARCH" = "x86_64" ]; then MIGRATE_ARCH="amd64"; elif [ "$ARCH" = "aarch64" ]; then MIGRATE_ARCH="arm64"; else MIGRATE_ARCH="amd64"; fi && \
    curl -L "https://github.com/golang-migrate/migrate/releases/download/v4.19.1/migrate.linux-${MIGRATE_ARCH}.tar.gz" | tar -xz -C /tmp && \
    mv /tmp/migrate /go/bin/migrate && \
    rm -rf /tmp/*

# Copy source and build binaries
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    mkdir -p bin && \
    go build -ldflags="-s -w" -o bin/api cmd/api/main.go && \
    go build -ldflags="-s -w" -o bin/seeder cmd/seeder/main.go && \
    go build -ldflags="-s -w" -o bin/gorm-migrate cmd/migrate/main.go
RUN cp /go/bin/migrate ./migrate

# Stage 2: Final image
FROM alpine:latest

WORKDIR /app

# Install certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Set timezone to WIB (Asia/Jakarta)
ENV TZ=Asia/Jakarta

# Copy the compiled binary
COPY --from=builder /app/bin/api .
COPY --from=builder /app/bin/seeder .
COPY --from=builder /app/bin/gorm-migrate .
COPY --from=builder /app/migrate .
COPY --from=builder /app/config/migrations ./config/migrations

# Credentials are mounted from the VPS at runtime; never bake them into an image.
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

EXPOSE 8081

CMD ["./entrypoint.sh"]
