FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/internal/db/migrations ./internal/db/migrations

EXPOSE 8080

CMD ["./main"]