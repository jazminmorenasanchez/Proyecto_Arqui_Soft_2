FROM golang:1.22-alpine AS build
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o users-api ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/users-api /app/users-api
ENV GIN_MODE=release
EXPOSE 8081
CMD ["/app/users-api"]
