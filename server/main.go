package main

import (
	"fmt"
	"log"
	"net"
	"os"
	pb "tages/service/proto"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

func main() {
	lis, err := net.Listen("tcp", port)

	//create files directory
	err := os.Mkdir("files", 0777)
	if err != nil {
		fmt.Println("Directry already exists")
	}

	if err != nil {
		log.Fatal("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterImageUploadServiceServer(s, &server{})

	log.Printf("Starting gRPC listener on port " + port)
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve: %v", err)
	}
}
