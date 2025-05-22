# Build stage
FROM golang:1.24.2-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Get version info from build args
ARG VERSION_MAJOR=1
ARG VERSION_MINOR=0
ARG VERSION_PATCH=0
ARG VERSION_PRE_RELEASE=
ARG VERSION_BUILD_METADATA=

# Determine Git commit and build time
RUN apk add --no-cache git
RUN BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
    && GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
    && GIT_DIRTY=$(git status --porcelain 2>/dev/null | wc -l) \
    && if [ "$GIT_DIRTY" -gt 0 ]; then DIRTY="true"; else DIRTY=""; fi \
    && CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X 'prayog-serviceability-service/pkg/version.Major=${VERSION_MAJOR}' \
    -X 'prayog-serviceability-service/pkg/version.Minor=${VERSION_MINOR}' \
    -X 'prayog-serviceability-service/pkg/version.Patch=${VERSION_PATCH}' \
    -X 'prayog-serviceability-service/pkg/version.PreRelease=${VERSION_PRE_RELEASE}' \
    -X 'prayog-serviceability-service/pkg/version.BuildMetadata=${VERSION_BUILD_METADATA}' \
    -X 'prayog-serviceability-service/pkg/version.Commit=${GIT_COMMIT}' \
    -X 'prayog-serviceability-service/pkg/version.BuildTime=${BUILD_TIME}' \
    -X 'prayog-serviceability-service/pkg/version.Dirty=${DIRTY}'" \
    -o serviceability ./cmd/api

# Final stage
FROM alpine:latest

# Add ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/serviceability .
COPY --from=builder /app/configs /root/configs

# Set labels with version information
LABEL org.opencontainers.image.version="${VERSION_MAJOR}.${VERSION_MINOR}.${VERSION_PATCH}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"
LABEL org.opencontainers.image.created="${BUILD_TIME}"

# Expose the application port
EXPOSE 8080

# Command to run the executable
CMD ["./serviceability"] 