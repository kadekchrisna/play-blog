package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/kadekchrisna/grpc-go/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthodID string             `bson:"authod_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func dataToBlogPb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthodID,
		Content:  data.Content,
		Title:    data.Title,
	}
}

func (s *server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("Stream Server List Blog..........")
	blogs, errFind := collection.Find(context.Background(), primitive.D{{}})
	if errFind != nil {
		return status.Error(
			codes.Internal,
			fmt.Sprintf("Error when finding. %v", errFind),
		)
	}
	defer blogs.Close(context.Background())
	for blogs.Next(context.Background()) {
		data := blogItem{}
		err := blogs.Decode(&data)
		if err != nil {
			return status.Error(
				codes.Internal,
				fmt.Sprintf("Error when Decoding. %v", errFind),
			)

		}
		stream.Send(&blogpb.ListBlogResponse{Blog: dataToBlogPb(&data)})
	}
	if err := blogs.Err(); err != nil {
		return status.Error(
			codes.Internal,
			fmt.Sprintf("Error when streaming. %v", errFind),
		)
	}
	return nil
}

func (s *server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Delete Blog..........")

	blogID := req.GetBlogId()
	iod, errID := primitive.ObjectIDFromHex(blogID)
	if errID != nil {
		return nil, status.Error(
			codes.Internal,
			fmt.Sprintf("Error when getting ID. %v", errID),
		)
	}
	filter := bson.M{"_id": iod}

	resDelete, errDelete := collection.DeleteOne(context.Background(), filter)
	if errDelete != nil {
		return nil, status.Error(
			codes.Internal,
			fmt.Sprintf("Error when Deleting. %v", errDelete),
		)

	}
	if resDelete.DeletedCount == 0 {
		return nil, status.Error(
			codes.NotFound,
			fmt.Sprintf("No blog found with id %v", iod),
		)

	}
	return &blogpb.DeleteBlogResponse{BlogId: iod.Hex()}, nil

}

func (s *server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("Update Blog..........")

	blog := req.GetBlog()
	iod, errID := primitive.ObjectIDFromHex(blog.Id)
	if errID != nil {
		return nil, status.Error(
			codes.Internal,
			fmt.Sprintf("Error when getting ID. %v", errID),
		)
	}
	data := blogItem{}

	filter := bson.M{"_id": iod}
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(&data); err != nil {
		return nil, status.Error(
			codes.Internal,
			fmt.Sprintf("Error when Decode. %v", err),
		)

	}

	data.AuthodID = blog.AuthorId
	data.Content = blog.Content
	data.Title = blog.Title

	_, errUpdate := collection.ReplaceOne(context.Background(), filter, data)
	if errUpdate != nil {
		return nil, status.Error(
			codes.Internal,
			fmt.Sprintf("Error when ReplaceOne. %v", errUpdate),
		)
	}
	return &blogpb.UpdateBlogResponse{Blog: dataToBlogPb(&data)}, nil

}

func (s *server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Read Blog..........")
	blogID := req.GetBlogId()
	iod, errID := primitive.ObjectIDFromHex(blogID)
	if errID != nil {
		return nil, status.Error(
			codes.Internal,
			fmt.Sprintf("Error when getting ID. %v", errID),
		)
	}

	data := blogItem{}

	filter := bson.M{"_id": iod, "authod_id": "K"}
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(&data); err != nil {
		return nil, status.Error(
			codes.Internal,
			fmt.Sprintf("Error when Decode. %v", err),
		)

	}

	return &blogpb.ReadBlogResponse{Blog: dataToBlogPb(&data)}, nil
}

func (s *server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Create Blog..........")
	blog := req.GetBlog()
	data := blogItem{
		AuthodID: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}
	res, errInsert := collection.InsertOne(context.Background(), data)
	if errInsert != nil {
		return nil, status.Error(
			codes.Internal,
			fmt.Sprintf("Error when inserting. %v", errInsert),
		)
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Error(
			codes.Internal,
			fmt.Sprintf("Error when getting latest ID. %v", errInsert),
		)

	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: data.AuthodID,
			Content:  data.Content,
			Title:    data.Title,
		},
	}, nil

}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	client, errNewClient := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if errNewClient != nil {
		log.Fatalf("NewClient err &v", errNewClient)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	errConnect := client.Connect(ctx)
	if errConnect != nil {
		log.Fatalf("Connecting err &v", errConnect)
	}
	collection = client.Database("goblog-v1").Collection("blog")

	p := ":50051"

	fmt.Println("Server starting.....")

	l, err := net.Listen("tcp", p)
	if err != nil {
		log.Fatalf("Listening err &v", err)
	}
	s := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(s, &server{})

	reflection.Register(s)

	go func() {
		if errServe := s.Serve(l); errServe != nil {
			log.Fatalf("Listening err &v", errServe)
		}
	}()
	fmt.Println("Server started on ", p)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
	fmt.Println("Stopping server....")
	s.Stop()
	fmt.Println("Closing listener...")
	l.Close()
	fmt.Println("Disconnecting DB....")
	client.Disconnect(context.TODO())
	fmt.Println("End.")

}
