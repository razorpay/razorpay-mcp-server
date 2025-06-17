FROM golang:1.24.2-alpine AS builder

# Install git
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build arguments with defaults
ARG VERSION="dev"
ARG COMMIT
ARG BUILD_DATE

# Use build args if provided, otherwise use fallbacks
RUN if [ -z "$COMMIT" ]; then \
        COMMIT=$(git rev-parse HEAD 2>/dev/null || echo 'unknown'); \
    fi && \
    if [ -z "$BUILD_DATE" ]; then \
        BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ); \
    fi && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}" -o razorpay-mcp-server ./cmd/razorpay-mcp-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Create a non-root user to run the application
RUN addgroup -S rzpgroup && adduser -S rzp -G rzpgroup

WORKDIR /app

COPY --from=builder /app/razorpay-mcp-server .

# Change ownership of the application to the non-root user
RUN chown -R rzp:rzpgroup /app

ENV CONFIG="" \
    RAZORPAY_KEY_ID="" \
    RAZORPAY_KEY_SECRET="" \
    PORT="8090" \
    MODE="stdio" \
    LOG_FILE=""

# Switch to the non-root user
USER rzp

# Use shell form to allow variable substitution and conditional execution
ENTRYPOINT ["sh", "-c", "./razorpay-mcp-server stdio --key ${RAZORPAY_KEY_ID} --secret ${RAZORPAY_KEY_SECRET} ${CONFIG:+--config ${CONFIG}} ${LOG_FILE:+--log-file ${LOG_FILE}}"]