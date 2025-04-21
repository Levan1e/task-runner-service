FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o runner .



FROM alpine:3.17

WORKDIR /app

COPY --from=builder /app/runner .

COPY --from=builder /app/internal/config/config.yaml ./internal/config/config.yaml

ENV REDIS_ADDR=redis:6379
ENV REDIS_PASSWORD=""
ENV REDIS_DB=0

EXPOSE 8080

ENTRYPOINT ["./runner"]
