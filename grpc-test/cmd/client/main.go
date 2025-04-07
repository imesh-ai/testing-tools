package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	msg "imesh.ai/grpc-test/messaging"
)

const (
	ENV_ADDRESS  = "SERVER_ADDRESS"
	ENV_MESSAGE  = "CLIENT_MESSAGE"
	ENV_INTERVAL = "CLIENT_INTERVAL"
	ENV_TIMEOUT  = "CLIENT_TIMEOUT"
	ENV_MAXREQ   = "CLIENT_MAX_REQ"
)

var (
	addr     string = "localhost:8080"
	message  string = "hello from client"
	interval string = "1s"
	timeout  uint   = 1
	maxReq   uint   = 1
)

func parseEnv() error {
	envAddr := os.Getenv(ENV_ADDRESS)
	envMessage := os.Getenv(ENV_MESSAGE)
	envInterval := os.Getenv(ENV_INTERVAL)
	envTimeout := os.Getenv(ENV_TIMEOUT)
	envMaxreq := os.Getenv(ENV_MAXREQ)

	if envAddr != "" {
		addr = envAddr
	}

	if envMessage != "" {
		message = envMessage
	}

	if envInterval != "" {
		interval = envInterval
	}

	if envTimeout != "" {
		timeoutUint, err := strconv.ParseUint(envTimeout, 10, 16)
		if err != nil {
			return err
		}

		timeout = uint(timeoutUint)
	}

	if envMaxreq != "" {
		maxReqUint, err := strconv.ParseUint(envMaxreq, 10, 16)
		if err != nil {
			return err
		}

		maxReq = uint(maxReqUint)
	}

	return nil
}

func sendRequest(
	conn msg.MessagingClient,
	ctx context.Context,
) {
	r, err := conn.BasicRequestReply(ctx, &msg.BasicMessage{Message: message})
	if err != nil {
		log.Fatalf("could not send message: %v", err)
	}
	log.Printf("Received a message: %s", r.GetMessage())
}

func shouldSendRequest(i uint) bool {
	if maxReq == 0 {
		return true
	}
	return i < maxReq
}

func main() {
	err := parseEnv()
	if err != nil {
		log.Fatalf("failed parsing environment variables: %v", err)
	}

	log.Printf("Connecting to %s", addr)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := msg.NewMessagingClient(conn)

	duration, err := time.ParseDuration(interval)
	if err != nil {
		log.Fatalf("failed to parse given interval %s: %v", interval, err)
	}
	var i uint
	for i = 0; shouldSendRequest(i); i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		sendRequest(c, ctx)
		time.Sleep(duration)
	}
}
