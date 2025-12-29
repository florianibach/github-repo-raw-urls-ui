# Build-Stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o github-repo-raw-urls-ui ./...

# Runtime-Stage
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/github-repo-raw-urls-ui /app/github-repo-raw-urls-ui
COPY --from=builder /app/templates /app/templates

EXPOSE 8080

ENTRYPOINT ["/app/github-repo-raw-urls-ui"]
