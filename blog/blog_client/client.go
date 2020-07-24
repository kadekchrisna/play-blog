package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/kadekchrisna/grpc-go/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Client....")
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Connection failed &f", err)
	}

	defer conn.Close()

	c := blogpb.NewBlogServiceClient(conn)

	fmt.Println("\n Create Blog Client\n")

	blog := &blogpb.Blog{
		AuthorId: "K",
		Content:  "Definitely not another lame story.",
		Title:    "Not A Lame Story.",
	}
	res, errCreate := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if errCreate != nil {
		log.Fatalf("Unexpected error. %v", errCreate)
	}
	fmt.Println("\n blog created, ", res.GetBlog())

	fmt.Println("\n Read Blog Client\n")

	blogId := res.GetBlog().GetId()

	_, errRead1 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "asdkjaskjdhkajs"})
	if errRead1 != nil {
		fmt.Printf("Unexpected error. %v\n\n", errRead1)
	}

	resRead, errRead2 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: blogId})
	if errRead2 != nil {
		log.Fatalf("Unexpected error. %v", errRead2)
	}
	fmt.Println("\n blog read success, ", resRead.GetBlog())

	fmt.Println("\n Update Blog Client\n")

	updateBlog := &blogpb.Blog{
		Id:       blogId,
		AuthorId: "L",
		Content:  "Definitely another lame story.",
		Title:    "A Lame Story.",
	}
	resUpdate, errUpdate := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: updateBlog})
	if errUpdate != nil {
		log.Fatalf("Unexpected error. %v", errUpdate)
	}
	fmt.Println("\n blog update success, ", resUpdate.GetBlog())

	fmt.Println("\n Delete Blog Client\n")

	resDelete, errDelete := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: blogId})
	if errDelete != nil {
		log.Fatalf("Unexpected error. %v", errDelete)
	}
	fmt.Println("\n blog delete success, ", resDelete.GetBlogId())

	fmt.Println("\n Lists Stream Servers Blog Client\n")

	stream, errLists := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if errLists != nil {
		log.Fatalf("Lists Stream fail while getting stream: &v", errLists)
	}
	for {
		res, errstream := stream.Recv()
		if errstream == io.EOF {
			fmt.Println("Stream done")
			break
		}
		if errstream != nil {
			log.Fatalf("Lists Stream fail while streaming: &v", errstream)
		}
		fmt.Println("Result stream: ", res.GetBlog())
	}

}
