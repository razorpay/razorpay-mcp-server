FROM golang:1.24.2-alpine AS builder

# Install git
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG VERSION="dev"
ARG COMMIT=""
ARG BUILD_DATE=""

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo 'unknown')} -X main.date=${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}" -o razorpay-mcp-server ./cmd/razorpay-mcp-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Create a non-root user to run the application
RUN addgroup -S rzpgroup && adduser -S rzp -G rzpgroup

# Create logs directory
RUN mkdir -p /app/logs && chown -R rzp:rzpgroup /app

WORKDIR /app

COPY --from=builder /app/razorpay-mcp-server .

# Change ownership of the application to the non-root user
RUN chown -R rzp:rzpgroup /app

ENV CONFIG="" \
    RAZORPAY_KEY_ID="" \
    RAZORPAY_KEY_SECRET="" \
    PORT="" \
    MODE="stdio" \
    LOG_FILE="/app/logs" \
    ADDRESS="localhost" \
    TOOLSETS="" \
    READ_ONLY="false"

# Switch to the non-root user
USER rzp

# Expose ports for both SSE (8090) and streamable HTTP (8080)
EXPOSE 8080 8090

# Use shell form to allow variable substitution and conditional execution
ENTRYPOINT ["sh", "-c", "\
echo 'logs are stored in: /app/logs'; \
case \"$MODE\" in \
    \"sse\") \
        PORT_FLAG=\"${PORT:-8090}\"; \
        echo \"Razorpay MCP Server running on sse\"; \
        ./razorpay-mcp-server sse --port $PORT_FLAG --address ${ADDRESS} ${TOOLSETS:+--toolsets ${TOOLSETS}} ${READ_ONLY:+--read-only}; \
        ;; \
    \"streamable-http\") \
        PORT_FLAG=\"${PORT:-8080}\"; \
        echo \"Razorpay MCP Server running on streamable-http\"; \
        ./razorpay-mcp-server streamable-http --port $PORT_FLAG --address ${ADDRESS} ${TOOLSETS:+--toolsets ${TOOLSETS}} ${READ_ONLY:+--read-only}; \
        ;; \
    *) \
        echo \"Razorpay MCP Server running on stdio\"; \
        ./razorpay-mcp-server stdio --key ${RAZORPAY_KEY_ID} --secret ${RAZORPAY_KEY_SECRET} ${LOG_FILE:+--log-file ${LOG_FILE}} ${TOOLSETS:+--toolsets ${TOOLSETS}} ${READ_ONLY:+--read-only}; \
        ;; \
esac"]
