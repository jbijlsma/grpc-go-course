package main

import (
	"context"
	"fmt"
	"github.com/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("Error %v", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)

	//calculateSum(c)
	//calculateAverage(c)
	//calculateMax(c)

	calculateSqrt(c, 16)
	calculateSqrt(c, -1)
}

func calculateSqrt(c calculatorpb.CalculatorServiceClient, number float64){

	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{ Number: number, })
	if err != nil {
		grpcErr, ok := status.FromError(err)
		if ok {
			fmt.Println("Grpc error. Code:", grpcErr.Code(), "Message", grpcErr.Message())
			return
		} else {
			log.Fatalf("Non grpc error %v", err)
		}
	}

	fmt.Printf("Sqrt of %v is %v\n", number, res.GetSquareRoot())
}

func calculateMax(c calculatorpb.CalculatorServiceClient){

	stream, err := c.Max(context.Background())
	if err != nil {
		log.Fatalf("Error when opening client stream")
	}

	waitc := make(chan bool)

	go func() {
		defer stream.CloseSend()

		numbers := []int32{1,5,3,6,2,20}

		for _, number := range numbers {
			err := stream.Send(&calculatorpb.MaxRequest{ Number: number, })
			if err != nil {
				log.Fatalf("Error while sending on stream %v", err)
			}

			time.Sleep(time.Second)
		}
	}()

	go func() {
		defer close(waitc)

		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving from stream %v", err)
			}

			fmt.Println("New max", res.GetMax())
		}
	}()

	<-waitc
}

func calculateAverage(c calculatorpb.CalculatorServiceClient) {

	stream, err := c.Average(context.Background())
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	numbers := []int32{1,2,3,4}

	for _, number := range numbers {
		err := stream.Send(&calculatorpb.AverageRequest{ Number: number },)
		if err != nil {
			log.Fatalf("Error sending req %v", err)
		}

		time.Sleep(time.Second)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	fmt.Println(res.GetResult())
}

func calculateSum(c calculatorpb.CalculatorServiceClient) {

	req := calculatorpb.SumRequest{
		Numbers: []int32{2, 3},
	}

	res, err := c.Add(context.Background(), &req)
	if err != nil {
		log.Fatalf("Error %v", err)
	}

	fmt.Println("Sum is", res.GetResult())
}
