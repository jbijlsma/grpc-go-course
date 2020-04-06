package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"

	"github.com/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
)

type server struct {
}

func (server) Add(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {

	var sum int32 = 0

	for _, i := range req.GetNumbers() {
		sum += int32(i)
	}

	result := calculatorpb.SumResponse{
		Result: sum,
	}

	return &result, nil
}

func (*server) Average(stream calculatorpb.CalculatorService_AverageServer) error{

	n := 0
	var sum float64 = 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&calculatorpb.AverageResponse{
				Result: sum / float64(n),
			})
		}
		if err != nil {
			log.Fatalf("Error %v", err)
		}

		fmt.Println("Received", req.GetNumber())

		n++
		sum += float64(req.GetNumber())
	}
}

func (*server) Max(stream calculatorpb.CalculatorService_MaxServer) error {
	var max int32

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error readinging from stream %v", err)
			return err
		}

		number := req.GetNumber()
		if number > max {
			max = number
			stream.Send(&calculatorpb.MaxResponse{ Max: max, })
		}
	}
}

func (*server) SquareRoot(_ context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error){

	number := req.GetNumber()

	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot use a negative number %v", number))
	}

	sqrt := math.Sqrt(number)

	return &calculatorpb.SquareRootResponse{
		SquareRoot: sqrt,
	}, nil
}

func main() {
	fmt.Println("Calculator server started")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Unable to start listener %v", err)
	}

	s := grpc.NewServer()

	reflection.Register(s)

	srv := server{}

	calculatorpb.RegisterCalculatorServiceServer(s, &srv)

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("Unable to start server %v", err)
	}
}
