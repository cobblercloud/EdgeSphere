FROM golang:1.20-alpine AS builder
RUN apk add --no-cache gcc musl-dev linux-headers
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w" -o /edge-gateway ./cmd/edge-gateway

FROM alpine:3.18
RUN apk add --no-cache libc6-compat
COPY --from=builder /edge-gateway /app/edge-gateway
COPY --from=builder /app/config.yaml /app/
VOLUME /data

# 启用io_uring支持
ENV GODEBUG=asyncpreemptoff=1
CMD ["/app/edge-gateway"]