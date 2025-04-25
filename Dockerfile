FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o razorpay-mcp-server ./cmd/razorpay-mcp-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/razorpay-mcp-server .

ENV MODE=stdio \
    PORT=8080 \
    CONFIG="" \
    KEY="" \
    SECRET="" \
    PAYMENT_MODE=test \
    LOG_FILE=""

ENTRYPOINT ["sh", "-c", "./razorpay-mcp-server ${MODE} --port ${PORT} --key ${KEY} --secret ${SECRET} --mode ${PAYMENT_MODE} ${CONFIG:+--config ${CONFIG}} ${LOG_FILE:+--log-file ${LOG_FILE}}"]
