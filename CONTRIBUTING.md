# Contributing to exist.io

Thank you for your interest in contributing to exist.io. We welcome pull requests, issue reports, and documentation improvements from both human developers and autonomous AI coding agents.

## Core Principles
1. **Truth Over Marketing**: All features, capabilities, and compatibility claims must be supported by automated, reproducible tests.
2. **Mutation Verification**: A test must fail if the core logic it tests is removed. Avoid tautological tests or untestable mocks.
3. **Pure Go / Zero CGO**: The backend must compile with `CGO_ENABLED=0` across all architectures (`linux/amd64`, `linux/arm64`, `darwin/arm64`, `windows/amd64`).
4. **Clean Commits**: Follow Conventional Commits (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`). Keep `main` green at all times.

## Local Development & Testing

### Prerequisites
- Go 1.22+ (tested through Go 1.27)
- Node.js 20+ and npm 10+
- Git

### Backend Commands
```bash
# Compile server daemon and CLI
go build ./cmd/exist ./cmd/existd

# Run all unit and integration tests with data race detection
go test -race ./...

# Run code vet
go vet ./...
```

### Frontend Commands
```bash
cd web

# Install dependencies
npm install

# Run automated Vitest test suite
npm test

# Build production bundle
npm run build
```

## Pull Request Guidelines
1. Ensure `go test -race ./...`, `npm test`, and `npm run build` all pass cleanly with zero warnings or errors.
2. Include unit or integration tests for all new endpoints, tools, and bug fixes.
3. Keep pull requests focused on a single change or fix.
