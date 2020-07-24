package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/kadekchresna/grpc-go/greet/greetpb"
	"google.golang.org/grpc"
)

type server struct{}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	log.Printf("RPC Req Greet: &v", req)
	name := req.GetFirstName()
	result := "Eyyy " + name
	return &greetpb.GreetResponse{Greet: result}, nil
}

func (*server) GreetStream(req *greetpb.GreetStreamRequest, stream greetpb.GreetService_GreetStreamServer) error {
	log.Printf("RPC Req Greet Stream: &v", req)
	name := req.GetFirstName()
	for i := 0; i < 10; i++ {
		result := "Eyyy, " + name + " number " + strconv.Itoa(i)
		res := &greetpb.GreetStreamResponse{
			Greet: result,
		}
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)

	}
	return nil
}

func (*server) GreetClientStream(stream greetpb.GreetService_GreetClientStreamServer) error {
	log.Println("RPC Req Greet Client Stream")
	result := ""
	for {
		req, err := stream.Recv()
		log.Printf("RPC Req Greet Stream: &v", req)
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.GreetClientStreamResponse{
				Greet: result,
			})

		}
		if err != nil {
			log.Fatalf("Error occurred on GreetClientStream: &v", err)
			return err
		}
		firstName := req.GetFirstName()
		result += "YEET " + firstName
	}

}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	log.Println("RPC Req GreetEveryone Client Server Stream")

	for {
		request, errRecv := stream.Recv()
		if errRecv == io.EOF {
			return nil
		}
		if errRecv != nil {
			log.Fatalf("Error Occurred on GreetEveryone: &v", errRecv)
			return errRecv
		}
		name := request.GetFirstName()
		result := "Eyyy, " + name
		fmt.Println("Request: ", name)
		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{
			Greet: result,
		})
		if sendErr != nil {
			log.Fatalf("Error Occurred on GreetEveryone: &v", sendErr)
			return sendErr
		}

	}

}

func main() {
	p := ":50051"
	fmt.Println("Server starting.....")
	l, err := net.Listen("tcp", p)
	if err != nil {
		log.Fatalf("Listening err &v", err)
	}
	s := grpc.NewServer()
	greetpb.RegisterGreetServiceServer(s, &server{})

	if errServe := s.Serve(l); errServe != nil {
		log.Fatalf("Listening err &v", errServe)

	}
	fmt.Println("Server started on ", p)

}
