FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o razorpay-mcp-server ./cmd/razorpay-mcp-server

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
    LOG_FILE=""

# Switch to the non-root user
USER rzp

ENTRYPOINT ["sh", "-c", "./razorpay-mcp-server stdio --key ${RAZORPAY_KEY_ID} --secret ${RAZORPAY_KEY_SECRET} ${CONFIG:+--config ${CONFIG}} ${LOG_FILE:+--log-file ${LOG_FILE}}"]
