# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o posthog-proxy .

# Runtime stage
FROM alpine:3.19

RUN adduser -D -u 1000 appuser

WORKDIR /app

COPY --from=builder /app/posthog-proxy .

USER appuser

EXPOSE 8080

CMD ["./posthog-proxy"]
