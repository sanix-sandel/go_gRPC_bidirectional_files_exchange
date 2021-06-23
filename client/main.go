package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	pb "tages/client/proto"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func testUploadImage(imageClient pb.ImageUploadServiceClient) {
	uploadImage(imageClient, "tmp/javascript.png")
}

func uploadImage(imageClient pb.ImageUploadServiceClient, imagePath string) {
	file, err := os.Open(imagePath)

	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}

	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := imageClient.UploadImage(ctx)

	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}
	fil_n := strings.Split(imagePath, "/")
	filename := fil_n[len(fil_n)-1]
	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				Name:      filename,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)

	if err != nil {
		log.Fatal("cannot send image info to server: ", err, stream)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunck to buffer: ", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_Chunkdata{
				Chunkdata: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err)
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image uploaded with name: %s, size: %d", res.GetName(), res.GetSize())
}

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()

	c := pb.NewImageUploadServiceClient(conn)
	testUploadImage(c)

	//list all images
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.ListImages(ctx, &wrappers.StringValue{Value: ""})
	if err != nil {
		log.Println(err)
	}
	log.Println(r)
}
