package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
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
	uploadImage(imageClient, "tmp/python.png")
	uploadImage(imageClient, "tmp/scala.png")
}

func uploadImage(imageClient pb.ImageUploadServiceClient, imagePath string) {
	file, err := os.Open(imagePath)

	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}

	stats, err := file.Stat()
	if err != nil {
		fmt.Println(err)
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
				Name: filename,
				//ImageType: filepath.Ext(imagePath),
				Created:  stats.ModTime().String(),
				Modified: stats.ModTime().String(),
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

/*
func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}
*/
func Save(imageName string, imageData bytes.Buffer) (string, error) {
	filename := path.Join("files", imageName)
	//data := bufio.NewWriter(&imageData)
	err := ioutil.WriteFile(filename, imageData.Bytes(), 0777)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}
	return imageName, nil
}

func DownloadImage(imageClient pb.ImageUploadServiceClient, filename string) {

	/*req, err:=stream.Recv()
	if err!=nil{
		return
	}*/

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := imageClient.DownloadImage(ctx, &wrappers.StringValue{Value: filename})
	if err != nil {
		fmt.Println(err)
	}

	res, err := stream.Recv()
	if err != nil {
		fmt.Println("cannot receive image info")
	}
	fmt.Println(res)
	///all image info

	imageData := bytes.Buffer{}

	for {
		res, err := stream.Recv()
		chunk := res.GetChunkdata()
		size := len(chunk)

		log.Println("chunck of %d received", size)

		if err == io.EOF {
			log.Println("No more date to receive")
			break
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			fmt.Printf("cannot write chunk data: %v", err)
		}
	}

	imageName, err := Save(filename, imageData)
	if err != nil {
		fmt.Printf("cannot save image to the store: %v", err)
	}

	fmt.Printf("Downloaded image with name %s", imageName)

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

	DownloadImage(c, "java.jpg")
	DownloadImage(c, "Rust.jpg")
	//download an image

}
