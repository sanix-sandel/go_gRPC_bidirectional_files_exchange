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
	pb "tages/service/proto"

	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

func (s *server) Save(imageName string, imageData bytes.Buffer) (string, error) {

	filename := path.Join("files", imageName)
	//data := bufio.NewWriter(&imageData)
	err := ioutil.WriteFile(filename, imageData.Bytes(), 0777)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}
	return imageName, nil
}

func (s *server) UploadImage(stream pb.ImageUploadService_UploadImageServer) error {
	//req, err:=stream.Recv()
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info"))
	}
	imageName := req.GetInfo().GetName()

	fmt.Println(imageName)

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		req, err := stream.Recv()
		chunk := req.GetChunkdata()
		size := len(chunk)

		log.Printf("Chunck of %d received", size)
		imageSize += size

		if err == io.EOF {
			//return stream.SendAndClose()
			log.Print("No more data")
			break
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return logError(status.Errorf(codes.Internal, "cannot write chunk data: %v", err))

		}
	}

	//create and save the image here
	imageName, err = s.Save(imageName, imageData)
	if err != nil {
		return logError(status.Errorf(codes.Internal, "cannot save image to the store: %v", err))
	}

	res := &pb.UploadImageResponse{
		Name: imageName,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)

	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot send response: %v", err))
	}

	log.Printf("saved image with id: %s, size: %d", imageName, imageSize)
	return nil
}

func (s *server) ListImages(ctx context.Context, message *wrappers.StringValue) (*pb.ImageList, error) {

	liste := []*pb.ImageInfo{}

	files, err := ioutil.ReadDir("./files")
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {

		//stats, err := os.Stat(f)
		if err != nil {
			fmt.Println(err)
		}

		file := &pb.ImageInfo{Name: f.Name(), Created: f.ModTime().String(), Modified: f.ModTime().String()} //f.ModTime()
		//log.Println("filename : ", stats.Name())
		//log.Println("file size : ", stats.Size())
		//log.Println("file date : ", stats.ModTime())
		liste = append(liste, file)
	}

	images := &pb.ImageList{Images: liste}

	return images, status.New(codes.OK, "").Err()
}

func (s *server) DownloadImage(filename *wrappers.StringValue, stream pb.ImageUploadService_DownloadImageServer) error {

	//find file in the repository
	imagePath := "files/" + filename.Value
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	stats, err := file.Stat()
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()

	res := &pb.DownloadImageResponse{
		Data: &pb.DownloadImageResponse_Info{
			Info: &pb.ImageInfo{
				Name:     filename.Value,
				Created:  stats.ModTime().String(),
				Modified: stats.ModTime().String(),
			},
		},
	}

	err = stream.Send(res)

	if err != nil {
		log.Fatal("cannot send file to the client: ", err, stream)
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

		res := &pb.DownloadImageResponse{
			Data: &pb.DownloadImageResponse_Chunkdata{
				Chunkdata: buffer[:n],
			},
		}

		err = stream.Send(res)
		if err != nil {
			log.Fatal("cannot send chunk to the client: ", err)
		}
	}

	//res1, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image served with name: %s ", filename.Value)

	return nil

}
