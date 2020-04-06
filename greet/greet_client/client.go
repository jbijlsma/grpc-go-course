package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"sync"
	"time"

	"github.com/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Client running")

	useTls := false
	var opts = grpc.WithInsecure()

	if useTls {
		certFile := "ssl/ca.crt"
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		if err != nil {
			log.Fatalf("Error when reading client certs %v", err)
		}

		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Could not connect %v", err)
	}

	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)

	var wg sync.WaitGroup
	wg.Add(3)

	done := make(chan bool)

	go doUnary(
		c,
		5 * time.Second,
		"Jeroen",
		func(cancel func()) {
			time.Sleep(time.Second * 2)
			cancel()
		},
		done)

	go doUnary(
		c,
		1 * time.Second,
		"Marc",
		func(cancel func()) {
		},
		done)

	go doUnary(
		c,
		5 * time.Second,
		"Jan",
		func(cancel func()) {
		},
		done)

	go func() {
		for {
			select {
				case _, ok := <-done:
					if ok {
						wg.Done()
					}
			}
		}
	}()

	wg.Wait()

	close(done)

	//doServerStreaming(c)

	//doClientStreaming(c)

	//doBiDiStreaming(c)
}

func doUnary(c greetpb.GreetServiceClient, timeout time.Duration, firstName string, timeoutDelegate func(func()), done chan<- bool) {

	defer func() { done<-true }()

	req := greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: firstName,
			LastName:  "Bijlsma",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	go timeoutDelegate(cancel)
	defer cancel()

	res, err := c.Greet(ctx, &req)
	if err != nil {
		grpcErr, ok := status.FromError(err)
		if ok {
			if grpcErr.Code() == codes.DeadlineExceeded {
				fmt.Printf("Deadline exceeded %v\n", firstName)
			}

			if grpcErr.Code() == codes.Canceled {
				fmt.Printf("Canceled by client %v\n", firstName)
			}

			return
		} else {
			log.Fatalf("Error %v %v\n", firstName, err)
		}
	}

	log.Printf("Response from greet %v: %v", firstName, res.Result)
}

func doBiDiStreaming(c greetpb.GreetServiceClient){

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while starting streaming %v", err)
	}

	var waitc = make(chan bool)

	go GreetReceive(stream, waitc)

	func() {
		var wg sync.WaitGroup
		wg.Add(2)

		go GreetJeroen(stream, &wg)
		go GreetYenni(stream, &wg)

		wg.Wait()

		stream.CloseSend()
	}()


	<-waitc
}

func GreetReceive(stream greetpb.GreetService_GreetEveryoneClient, waitc chan<- bool){
	defer close(waitc)

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error receivning from stream %v", err)
		}

		fmt.Println(res.Result)
	}
}

func GreetJeroen(stream greetpb.GreetService_GreetEveryoneClient, wg *sync.WaitGroup) {
	defer wg.Done()

	for i:=0; i<5; i++ {
		stream.Send(&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName:            "Jeroen",
				LastName:             "Bijlsma",
			},
		})

		time.Sleep(time.Second)
	}
}

func GreetYenni(stream greetpb.GreetService_GreetEveryoneClient, wg *sync.WaitGroup) {
	defer wg.Done()

	for i:=0; i<5; i++ {
		stream.Send(&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName:            "Yenni",
				LastName:             "Widjaja",
			},
		})

		time.Sleep(time.Second)
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	requests := []greetpb.LongStreamRequest{
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Jeroen",
				LastName:  "Bijlsma",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Yenni",
				LastName:  "Widjaja",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Jan",
				LastName:  "Bijlsma",
			},
		},
		{
			Greeting: &greetpb.Greeting{
				FirstName: "Marjo",
				LastName:  "Bijlsma",
			},
		},
	}

	for _, req := range requests {
		err = stream.Send(&req)
		if err != nil {
			log.Fatalf("Error %v", err)
		}

		time.Sleep(time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	fmt.Println(res.GetResult())
}

func doServerStreaming(c greetpb.GreetServiceClient) {

	req := greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Jeroen",
			LastName:  "Bijlsma",
		},
	}

	resStream, err := c.GreetManyTimes(context.Background(), &req)
	if err != nil {
		log.Fatalf("Err %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Err %v", err)
		}

		fmt.Println(msg.GetResult())
	}
}
