package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"google.golang.org/grpc/testdata"

	pb "thechee/grpcAWS_test/protos" // set your path

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	tls      = flag.Bool("tls", true, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("../key/certs/mycert.pem", "../key/certs/mycert.pem", "The TLS cert file")
	keyFile  = flag.String("../key/private/mykey.pem", "../key/private/mykey.pem", "The TLS key file")
	//jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port = flag.Int("port", 8080, "The server port")
	svc  = &dynamodb.DynamoDB{}
)

// CheeTestServer class
type CheeTestServer struct {
	//...
}

// CheeObject object
type CheeObject struct {
	Data      string
	Timestamp int64
}

// CreateObject request response
// client request, server response
// request to insert a string with current timestamp, return Success or Fail
func (s *CheeTestServer) CreateObject(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	// Creating Object
	object := CheeObject{
		Data:      request.Message,
		Timestamp: time.Now().Unix(),
	}
	// Transform Object
	ob, err := dynamodbattribute.MarshalMap(object)
	if err != nil {
		fmt.Println("Got error marshalling new object:")
		fmt.Println(err.Error())
		return &pb.HelloReply{Message: "Error marshalling " + request.Message}, nil
	}
	//Put object to dynamoDB
	input := &dynamodb.PutItemInput{
		Item:      ob,
		TableName: aws.String("Mytable"),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		return &pb.HelloReply{Message: "Error putting " + request.Message}, nil
	}
	return &pb.HelloReply{Message: "Success " + request.Message}, nil
}

// GetObject request response
// client request, server response
// request to find a string, return to timestamp
func (s *CheeTestServer) GetObject(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	// Get Object
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("Mytable"),
		Key: map[string]*dynamodb.AttributeValue{
			"Data": {
				S: aws.String(request.Message),
			},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return &pb.HelloReply{Message: "Error getting " + request.Message}, nil
	}
	// Tranform Object
	object := CheeObject{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &object)
	if err != nil {
		//panic(fmt.Sprintf("Failed to unmarshal Record, %v", err))
		return &pb.HelloReply{Message: "Fail unmarshal " + request.Message}, nil
	}

	if object.Data == "" {
		fmt.Println("Could not find ")
		return &pb.HelloReply{Message: "Not found " + request.Message}, nil
	}
	return &pb.HelloReply{Message: "Get " + strconv.FormatInt(object.Timestamp, 10)}, nil
}

// GetStream request stream
// client request, server stream (probably wont used)
func (s *CheeTestServer) GetStream(request *pb.HelloRequest, stream pb.MyService_GetStreamServer) error {
	preset := [5]string{"I", "gonna", "stand", "for", "you"}
	for _, feature := range preset {
		// var response *pb.HelloReply
		// response.Message = feature
		if err := stream.Send(&pb.HelloReply{
			Message: feature,
		}); err != nil {
			return err
		}
	}
	return nil
}

// GiveStream stream response
// client stream, server response (probably wont used)
func (s *CheeTestServer) GiveStream(stream pb.MyService_GiveStreamServer) error {
	var words []string
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.HelloReply{
				Message: fmt.Sprintf("you give %d word, which are ...", len(words)),
			})
		}
		if err != nil {
			return err
		}
		words = append(words, request.Message)
	}
}

// Chat stream stream
// bistream (probably wont used)
func (s *CheeTestServer) Chat(stream pb.MyService_ChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		// var out *pb.HelloReply
		// out.Message = fmt.Sprintf("Receive %s", in.Name)
		if err := stream.Send(&pb.HelloReply{
			Message: "Hello " + fmt.Sprintf("Receive %s", in.Message),
		}); err != nil {
			return err
		}
	}
}

func newServer() *CheeTestServer {
	return &CheeTestServer{}
}

func main() {
	print("Start\n")
	flag.Parse()
	print("1\n")
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	print("2\n")
	var opts []grpc.ServerOption
	if *tls {
		if *certFile == "" {
			println("CERT not found use Alt")
			*certFile = testdata.Path("server1.pem")
		}
		if *keyFile == "" {
			*keyFile = testdata.Path("server1.key")
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	print("3\n")
	grpcServer := grpc.NewServer(opts...)
	print("4\n")
	pb.RegisterMyServiceServer(grpcServer, newServer())
	print("5\n")

	// AWS
	cfg := aws.Config{}
	cfg.Region = aws.String("eu-west-2")
	cfg.Endpoint = aws.String("http://localhost:10001")
	sess := session.Must(session.NewSession(&cfg))
	svc = dynamodb.New(sess)

	// AWS Table Creation
	fmt.Printf("Create Table...:\n")
	tableName := "Mytable"
	input := &dynamodb.CreateTableInput{
		AttributeDefinitions: []*dynamodb.AttributeDefinition{
			{
				AttributeName: aws.String("Data"),
				AttributeType: aws.String("S"),
			}, // multiple key available
		},
		KeySchema: []*dynamodb.KeySchemaElement{
			{
				AttributeName: aws.String("Data"),
				KeyType:       aws.String("HASH"),
			},
		},
		ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(10),
			WriteCapacityUnits: aws.Int64(10),
		},
		TableName: aws.String(tableName),
	}

	_, err = svc.CreateTable(input)
	if err != nil {
		fmt.Println("Got error calling CreateTable:")
		fmt.Println(err.Error())
		fmt.Println("SKIP!!!")
	}

	print("Done AWS\n")
	grpcServer.Serve(lis)
	print("Running...\n")
}
