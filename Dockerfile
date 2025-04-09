FROM golang:1.23.1 as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth ./cmd/auth/

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /auth .

CMD ["/root/auth"]