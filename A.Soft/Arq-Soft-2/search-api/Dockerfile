FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o search-api ./cmd/api

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/search-api /app/search-api
EXPOSE 8083
ENTRYPOINT ["/app/search-api"]
