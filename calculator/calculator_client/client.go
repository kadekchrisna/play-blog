package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/kadekchresna/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Connecting to server.....")
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error occurred on connecting to server: &v", err)
	}

	defer conn.Close()

	c := calculatorpb.NewCalculateSumClient(conn)
	// doCalcutale(c)
	// doStreamPrime(c)
	// doStreamAverage(c)
	doFinMax(c)

}

func doCalcutale(c calculatorpb.CalculateSumClient) {
	fmt.Println("doCalcutale")
	reqCalculate := calculatorpb.CalculatorRequest{
		Number: []int32{2, 3, 5},
	}
	result, errCalculate := c.Calculate(context.Background(), &reqCalculate)
	if errCalculate != nil {
		log.Fatalf("Error occurred on calculate: &v", errCalculate)
	}
	fmt.Println("Result: ")
	fmt.Println(result)
}

func doStreamPrime(c calculatorpb.CalculateSumClient) {
	fmt.Println("doCalcutale")

	req := calculatorpb.PrimeRequest{
		Number: 200,
	}
	stream, errPrimeStream := c.PrimeStream(context.Background(), &req)
	if errPrimeStream != nil {
		log.Fatalf("Error occurred on doStreamPrime: &v", errPrimeStream)
	}
	for {
		res, errstream := stream.Recv()
		if errstream == io.EOF {
			fmt.Println("Stream done")
			break
		}
		if errstream != nil {
			log.Fatalf("GreetStream stream fail: &v", errstream)
		}
		fmt.Println("Result stream: ", res)

	}

}

func doStreamAverage(c calculatorpb.CalculateSumClient) {
	fmt.Println("doStreamAverage")
	requests := []*calculatorpb.AverageRequest{
		&calculatorpb.AverageRequest{
			Number: 20,
		},
		&calculatorpb.AverageRequest{
			Number: 40,
		},
		&calculatorpb.AverageRequest{
			Number: 80,
		},
		&calculatorpb.AverageRequest{
			Number: 100,
		},
	}
	stream, errPrimeStream := c.AverageStream(context.Background())
	if errPrimeStream != nil {
		log.Fatalf("Error Occurred doStreamAverage: &v", errPrimeStream)
	}
	for _, val := range requests {
		fmt.Println("Stream value ", val)
		stream.Send(val)
	}
	res, errCloseAndRecv := stream.CloseAndRecv()
	if errCloseAndRecv != nil {
		log.Fatalf("Error Occurred doStreamAverage: &v", errCloseAndRecv)
	}
	fmt.Println(res)

}

func doFinMax(c calculatorpb.CalculateSumClient) {
	fmt.Println("doFinMax")
	requests := []*calculatorpb.FindMaxRequest{
		&calculatorpb.FindMaxRequest{
			Number: 20,
		},
		&calculatorpb.FindMaxRequest{
			Number: 2,
		},
		&calculatorpb.FindMaxRequest{
			Number: 1,
		},
		&calculatorpb.FindMaxRequest{
			Number: 20,
		},
	}
	stream, errFindMax := c.FindMax(context.Background())
	if errFindMax != nil {
		log.Fatalf("Error Occurred FindMax: &v", errFindMax)
	}

	ch := make(chan struct{})

	go func() {
		for _, val := range requests {
			fmt.Println("Request: ", val)
			errSend := stream.Send(val)
			if errSend != nil {
				log.Fatalf("Error Occurred FindMax Send: &v", errSend)
			}
			time.Sleep(1000 * time.Millisecond)
		}
		errClode := stream.CloseSend()
		if errClode != nil {
			log.Fatalf("Error Occurred FindMax CloseSend: &v", errClode)
		}
	}()

	go func() {
		for {
			res, errRes := stream.Recv()
			if errRes == io.EOF {
				break
			}
			if errRes != nil {
				log.Fatalf("Error Occurred FindMax Recv: &v", errRes)
			}
			fmt.Println(res)
		}
		close(ch)
	}()
	<-ch
}
