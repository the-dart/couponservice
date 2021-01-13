package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

// taken from https://tutorialedge.net/golang/go-grpc-beginners-tutorial/
func main() {

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}
