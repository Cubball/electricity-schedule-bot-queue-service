FROM golang:1.23 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/queue-service/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/main .
CMD ["./main"]
