FROM golang:1.24-bookworm AS builder

# Set necessary environment variables
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

RUN apt-get update && apt-get install -y \
    librdkafka-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
 
#copy source code
COPY  . .

RUN go build -o /dist/main ./cmd/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    librdkafka1 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /dist/main /

COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

EXPOSE 3000

ENTRYPOINT ["/main"]
