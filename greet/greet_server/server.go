package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"net"
	"time"

	"github.com/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
)

type server struct {
}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {

	var res greetpb.GreetResponse

	cancel := make(chan error)
	done := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(cancel)
				return
			}
		}
	}()

	go func() {
		defer close(done)

		time.Sleep(time.Second * 3)

		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName

		res = greetpb.GreetResponse{
			Result: result,
		}
	}()

	for {
		select {
			case <- cancel:
				return nil, status.Error(codes.Canceled, "Request was canceled")
			case <-done:
				return &res, nil
		}
	}
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {

	fmt.Printf("Greet function was invoked with %v", req)

	firstName := req.GetGreeting().GetFirstName()

	for i := 0; i < 10; i++ {
		res := greetpb.GreetManyTimesResponse{
			Result: "Hello " + firstName,
		}

		err := stream.Send(&res)
		if err != nil {
			log.Fatalf("Error %v", err)
			return nil
		}

		time.Sleep(time.Second)
	}

	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {

	res := ""

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongStreamResponse{
				Result:res,
			})
		}
		if err != nil {
			log.Fatalf("Error %v", err)
			return err
		}

		res += "Hi " + req.GetGreeting().FirstName + "\n"
	}
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error{
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while receiving from stream %v", err)
			return err
		}

		fmt.Printf("Receiced request %v\n", req)

		err = stream.Send(&greetpb.GreetEveryoneResponse{
			Result: "Hi " + req.GetGreeting().FirstName,
		})
		if err != nil {
			log.Fatalf("Error while sending to stream %v", err)
			return err
		}
	}
}

func main() {

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	fmt.Println("Server running")

	useTls := false
	var options []grpc.ServerOption

	if useTls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed loading certs %v", err)
		}

		options = append(options, grpc.Creds(creds))
	}
	s := grpc.NewServer(options...)

	reflection.Register(s)

	greetpb.RegisterGreetServiceServer(s, &server{})

	if err = s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}
