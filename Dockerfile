FROM golang:1.23-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . .

RUN go mod download && go build -o /app/bin/auth ./cmd/auth/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/auth /app/auth
COPY --from=builder /app/config/config.example.yaml /app/config/config.yaml
COPY --from=builder /app/db/init.sql /app/db/init.sql

WORKDIR /app

EXPOSE 8080

CMD ["/app/auth"]