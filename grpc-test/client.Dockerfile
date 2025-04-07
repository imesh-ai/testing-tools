# syntax=docker/dockerfile:1
FROM golang:1.22

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/client/main.go ./
RUN mkdir ./messaging
ADD messaging ./messaging

RUN CGO_ENABLED=0 GOOS=linux go build -o /client

ENV SERVER_ADDRESS  "localhost:8080"
ENV CLIENT_MESSAGE  "hello from client"
ENV CLIENT_INTERVAL "1s"
ENV CLIENT_TIMEOUT  "1"
ENV CLIENT_MAX_REQ  "1"

CMD ["/client"]