package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/kadekchresna/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
)

type (
	server struct{}
)

func (*server) Calculate(ctx context.Context, req *calculatorpb.CalculatorRequest) (*calculatorpb.CalculatorResponse, error) {
	fmt.Println("Calculate: ", req)
	numbers := req.GetNumber()
	if len(numbers) < 1 {
		return nil, errors.New("Number must contain 2 or move value")
	}
	var res int32
	for _, val := range numbers {
		res += val
	}

	result := &calculatorpb.CalculatorResponse{
		Number: res,
	}

	return result, nil

}

func (*server) PrimeStream(req *calculatorpb.PrimeRequest, stream calculatorpb.CalculateSum_PrimeStreamServer) error {
	fmt.Println("PrimeStream: ", req)

	number := req.GetNumber()

	var k int32
	for k = 2; ; {
		if number%k == 0 {
			res := &calculatorpb.PrimeResponse{
				Number: k,
			}
			stream.Send(res)
			number = number / k
			fmt.Println(number)
			time.Sleep(1000 * time.Millisecond)
			if number == 1 {
				break
			}
		} else {
			k++
		}
	}
	return nil

}

func (*server) AverageStream(stream calculatorpb.CalculateSum_AverageStreamServer) error {
	fmt.Println("AverageStream")
	var result float32
	var length int
	for {
		req, errRecv := stream.Recv()
		if errRecv == io.EOF {
			result = result / float32(length)
			return stream.SendAndClose(&calculatorpb.AverageResponse{
				Number: result,
			})

		}
		if errRecv != nil {
			log.Fatalf("Error Occurred on AverageStream: &v", errRecv)
		}
		result += float32(req.GetNumber())
		length++
	}

}

func (*server) FindMax(stream calculatorpb.CalculateSum_FindMaxServer) error {
	fmt.Println("FindMax")
	var number int32
	var lowest int32
	for {
		req, errRecv := stream.Recv()
		if errRecv == io.EOF {
			return nil
		}
		if errRecv != nil {
			log.Fatalf("Error Occurred on FindMax: &v", errRecv)
			return errRecv
		}
		fmt.Println("Request value: ", req.GetNumber())
		number = req.GetNumber()
		if lowest == 0 {
			lowest = number
		} else if number <= lowest {
			errSend := stream.Send(&calculatorpb.FindMaxResponse{
				Number: lowest,
			})
			lowest = number
			if errSend != nil {
				log.Fatalf("Error Occurred on Send: &v", errSend)
				return errSend
			}
		} else {
			errSend := stream.Send(&calculatorpb.FindMaxResponse{
				Number: number,
			})
			if errSend != nil {
				log.Fatalf("Error Occurred on Send: &v", errSend)
				return errSend
			}
		}
	}

}

func main() {
	p := ":50051"
	fmt.Println("Server starting.....")
	l, err := net.Listen("tcp", p)
	if err != nil {
		log.Fatalf("Error occurred on listening on &v &v", p, err)
	}
	s := grpc.NewServer()
	calculatorpb.RegisterCalculateSumServer(s, &server{})

	if errServe := s.Serve(l); errServe != nil {
		log.Fatalf("Error occurred on serving server: &v", errServe)
	}
}
