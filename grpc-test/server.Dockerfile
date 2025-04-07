# syntax=docker/dockerfile:1
FROM golang:1.22

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/server/main.go ./
RUN mkdir ./messaging
ADD messaging ./messaging

RUN CGO_ENABLED=0 GOOS=linux go build -o /server
EXPOSE 8080

ENV HOST  "localhost"
ENV PORT  "8080"
ENV REPLY "hello from server"

CMD ["/server"]
