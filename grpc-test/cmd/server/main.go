package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
	msg "imesh.ai/grpc-test/messaging"
)

const (
	ENV_HOST  = "HOST"
	ENV_PORT  = "PORT"
	ENV_REPLY = "REPLY"
)

var (
	host  string = "localhost"
	port  int    = 8080
	reply string = "hello from server"
)

func parseEnv() error {
	envHost := os.Getenv(ENV_HOST)
	envPort := os.Getenv(ENV_PORT)
	envReply := os.Getenv(ENV_REPLY)

	if envHost != "" {
		host = envHost
	}

	if envPort != "" {
		envPortInt, err := strconv.Atoi(envPort)
		if err != nil {
			return nil
		}

		port = envPortInt
	}

	if envReply != "" {
		reply = envReply
	}

	return nil
}

type server struct {
	msg.UnimplementedMessagingServer
}

func (s *server) BasicRequestReply(_ context.Context, in *msg.BasicMessage) (*msg.BasicMessage, error) {
	log.Printf("Received: %v", in.GetMessage())
	return &msg.BasicMessage{Message: reply}, nil
}

func main() {
	err := parseEnv()
	if err != nil {
		log.Fatalf("failed parsing environment variables: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	msg.RegisterMessagingServer(s, &server{})
	log.Printf("started listening at %v", lis.Addr())

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
