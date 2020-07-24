package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/kadekchresna/grpc-go/greet/greetpb"
	"google.golang.org/grpc"
)

var wg sync.WaitGroup

func main() {
	fmt.Println("Client....")
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Connection failed &f", err)
	}

	defer conn.Close()

	c := greetpb.NewGreetServiceClient(conn)
	// doGreet(c)
	// doStreamGreet(c)
	// doClientSream(c)
	doStreamClientServer(c)
	// fmt.Println(c)
}

func doGreet(c greetpb.GreetServiceClient) {
	fmt.Println("doGreet")
	req := greetpb.GreetRequest{
		FirstName: "K",
	}
	result, err := c.Greet(context.Background(), &req)
	if err != nil {
		log.Fatalf("Greet fail: &v", err)
	}
	fmt.Println(result)
}

func doStreamGreet(c greetpb.GreetServiceClient) {
	fmt.Println("doStreamGreet")

	req := greetpb.GreetStreamRequest{
		FirstName: "KKK",
	}

	stream, errGreetStream := c.GreetStream(context.Background(), &req)
	if errGreetStream != nil {
		log.Fatalf("GreetStream fail: &v", errGreetStream)
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

func doClientSream(c greetpb.GreetServiceClient) {
	fmt.Println("doClientSream")
	reqs := []*greetpb.GreetClientStreamRequest{
		&greetpb.GreetClientStreamRequest{
			FirstName: "k",
		},
		&greetpb.GreetClientStreamRequest{
			FirstName: "l",
		},
		&greetpb.GreetClientStreamRequest{
			FirstName: "m",
		},
		&greetpb.GreetClientStreamRequest{
			FirstName: "n",
		},
		&greetpb.GreetClientStreamRequest{
			FirstName: "o",
		},
	}
	stream, errGreetClientStream := c.GreetClientStream(context.Background())
	if errGreetClientStream != nil {
		log.Fatalf("GreetClientStream stream fail: &v", errGreetClientStream)
	}

	for _, val := range reqs {
		fmt.Println("Sending stream client &v", val)
		stream.Send(val)
		time.Sleep(1000 * time.Millisecond)
	}
	res, errCloseAndRecv := stream.CloseAndRecv()
	if errCloseAndRecv != nil {
		log.Fatalf("Error occurred on CloseAndRecv: &v", errCloseAndRecv)
	}
	fmt.Println(res)
}

func doStreamClientServer(c greetpb.GreetServiceClient) error {
	fmt.Println("doStreamClientServer")
	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			FirstName: "K",
		},
		&greetpb.GreetEveryoneRequest{
			FirstName: "l",
		},
		&greetpb.GreetEveryoneRequest{
			FirstName: "m",
		},
		&greetpb.GreetEveryoneRequest{
			FirstName: "n",
		},
		&greetpb.GreetEveryoneRequest{
			FirstName: "o",
		},
	}

	stream, errGreetEveryone := c.GreetEveryone(context.Background())
	if errGreetEveryone != nil {
		log.Fatalf("Error Occurred on GreetEveryone &v", errGreetEveryone)
		return errGreetEveryone
	}

	// go streamClientAsync(stream, requests, wg)
	// go streamServerAsync(stream, wg)

	ch := make(chan struct{})

	wg.Add(2)

	// make stream client by go routines
	go func() {
		for _, val := range requests {

			fmt.Println("Sending stream client &v", val)
			stream.Send(val)
			time.Sleep(1000 * time.Millisecond)

		}
		errCloseSend := stream.CloseSend()
		if errCloseSend != nil {
			log.Fatalf("Error Occurred on streamClientAsync &v", errCloseSend)
		}
		wg.Done()
	}()

	// recieve stream server by go routines
	go func() {
		for {
			res, errRecv := stream.Recv()
			if errRecv == io.EOF {
				break
			}
			// wg.Add(i)
			if errRecv != nil {
				log.Fatalf("Error Occurred on streamServerAsync &v", errRecv)

			}
			fmt.Println("Response : ", res)

		}
		close(ch)
		wg.Done()
	}()

	wg.Wait()

	// block until everything done
	// <-ch
	return nil

}

// func streamClientAsync(stream greetpb.GreetService_GreetEveryoneClient, requests []*greetpb.GreetEveryoneRequest, wg sync.WaitGroup) error {
// 	var i int
// 	for _, val := range requests {
// 		wg.Add(i)
// 		fmt.Println("Sending stream client &v", val)
// 		stream.Send(val)
// 		time.Sleep(1000 * time.Millisecond)
// 		i++
// 	}
// 	errCloseSend := stream.CloseSend()
// 	if errCloseSend != nil {
// 		log.Fatalf("Error Occurred on streamClientAsync &v", errCloseSend)
// 		return errCloseSend
// 	}

// 	return nil

// }

// func streamServerAsync(stream greetpb.GreetService_GreetEveryoneClient, wg sync.WaitGroup) error {
// 	var i int
// 	for {
// 		res, errRecv := stream.Recv()
// 		if errRecv == io.EOF {
// 			break
// 		}
// 		wg.Add(i)
// 		if errRecv != nil {
// 			log.Fatalf("Error Occurred on streamServerAsync &v", errRecv)
// 			return errRecv
// 		}
// 		fmt.Println("Response : ", res)
// 		i++
// 	}

// 	return nil
// }
