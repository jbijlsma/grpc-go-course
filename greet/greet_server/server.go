package main

import (
	"fmt"
	"log"
	"net"

	greetpb "github.com/grpc-go-course/greet/greetpb"
	grpc "google.golang.org/grpc"
)

type server struct {
}

greetpb.

func main() {

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	fmt.Println("Server running")

	s := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err = s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}
