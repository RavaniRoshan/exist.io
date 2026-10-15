# syntax=docker/dockerfile:1
# ------------------------------------------------------------------------------
# Stage 1: Build Enterprise Web UI Dashboard
# ------------------------------------------------------------------------------
FROM node:22-alpine AS web-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci --no-audit --prefer-offline
COPY web/ ./
RUN npm run build

# ------------------------------------------------------------------------------
# Stage 2: Build exist.io Enterprise Daemon and CLI
# ------------------------------------------------------------------------------
FROM golang:1.24-alpine AS go-builder
WORKDIR /src
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=v0.2.0-enterprise" -o /bin/existd ./cmd/existd && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -X main.version=v0.2.0-enterprise" -o /bin/exist ./cmd/exist

# ------------------------------------------------------------------------------
# Stage 3: Hardened, Non-Root Enterprise Production Container
# ------------------------------------------------------------------------------
FROM alpine:3.21

# Install runtime dependencies: git and git-http-backend are required for Smart-HTTP transport
RUN apk add --no-cache git ca-certificates tzdata \
    && rm -rf /var/cache/apk/*

# Create unprivileged service user and group
RUN addgroup -g 1000 exist && \
    adduser -u 1000 -G exist -s /bin/sh -D exist

# Create persistent storage directories with strict access permissions
RUN mkdir -p /var/lib/exist/git /var/lib/exist/audit && \
    chown -R exist:exist /var/lib/exist && \
    chmod 700 /var/lib/exist

# Copy binaries from builder
COPY --from=go-builder /bin/existd /usr/local/bin/existd
COPY --from=go-builder /bin/exist /usr/local/bin/exist

# Copy static Web UI assets into daemon directory
COPY --from=web-builder /app/web/dist /var/lib/exist/web/dist

USER exist:exist
WORKDIR /var/lib/exist

ENV EXIST_DATA_DIR=/var/lib/exist \
    EXIST_LISTEN_ADDR=0.0.0.0:8080 \
    EXIST_PUBLIC_URL=http://localhost:8080 \
    EXIST_GIT_URL=localhost:8080

EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/existd"]
