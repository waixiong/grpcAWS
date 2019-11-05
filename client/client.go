package main

import (
	"context"
	"flag"
	"log"
	"time"

	pb "thechee/grpcAWS_test/protos" // set your path

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

var (
	tls                = flag.Bool("tls", true, "Connection uses TLS if true, else plain TCP")
	caFile             = flag.String("ca_file", "../key/certs/mycert.pem", "The file containing the CA root cert file")
	serverAddr         = flag.String("server_addr", "localhost:8080", "The server address in the format of host:port")
	serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

// create something
func createObject(client pb.MyServiceClient, point *pb.HelloRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	response, err := client.CreateObject(ctx, point)
	if err != nil {
		log.Fatalf("%v.CreateObject(_) = _, %v: ", client, err)
	}
	log.Println(response)
}

// get that thing
func getObject(client pb.MyServiceClient, point *pb.HelloRequest) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	response, err := client.GetObject(ctx, point)
	if err != nil {
		log.Fatalf("%v.getObject(_) = _, %v: ", client, err)
	}
	log.Println(response)
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	if *tls {
		if *caFile == "" {
			print("cert not found\n")
			*caFile = testdata.Path("mykey.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	println("\tStage 1")
	opts = append(opts, grpc.WithBlock())
	println("\tStage 1.1")
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	println("\tStage 2")
	defer conn.Close()
	client := pb.NewMyServiceClient(conn)

	println("\tStart 1")
	// Looking for a valid feature
	createObject(client, &pb.HelloRequest{Message: "Try can?"})

	// Feature missing.
	getObject(client, &pb.HelloRequest{Message: "Try can?"})

	getObject(client, &pb.HelloRequest{Message: "Can found me??"})
}
