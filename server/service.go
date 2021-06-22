package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path"
	"sync"
	pb "tages/service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	filesMap DiskFileStore
	//filesList []*ImageInfo
}

//storage-related
type FileStore interface {
	Save(imageName string, imageType string, imageData bytes.Buffer)
}

type DiskFileStore struct {
	mutex       sync.RWMutex
	imageFolder string
	images      map[string]*ImageInfo
}

type ImageInfo struct {
	name string
	Type string
	Path string
}

func NewFileStore(imageFolder string) *DiskFileStore {
	return &DiskFileStore{
		imageFolder: imageFolder,
		images:      make(map[string]*ImageInfo),
	}
}

func logError(err error) error {
	if err != nil {
		log.Print(err)
	}
	return err
}

func (store *DiskFileStore) Save(imageName string, imageType string, imageData bytes.Buffer) (string, error) {
	/*imageID, err := uuid.NewRandom()
	  if err != nil {
	      return "", fmt.Errorf("cannot generate image id: %w", err)
	  }*/

	/*imagePath := fmt.Sprintf("%s/%s%s", store.imageFolder, imageName, imageType)

	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("cannot create image file: %w", err)
	}

	_, err = imageData.WriteTo(file)

	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}*/

	filename := path.Join("files", imageName)
	//data := bufio.NewWriter(&imageData)
	err := ioutil.WriteFile(filename, imageData.Bytes(), 0777)
	if err != nil {
		return "", fmt.Errorf("cannot write image to file: %w", err)
	}

	store.mutex.Lock()
	defer store.mutex.Unlock()

	/*store.images[imageName] = &ImageInfo{
		name: imageName,
		Type: imageType,
		Path: filename, //imagePath,
	}*/

	return imageName, nil
}

func (s *server) UploadImage(stream pb.ImageUploadService_UploadImageServer) error {
	//req, err:=stream.Recv()
	req, err := stream.Recv()
	if err != nil {
		return logError(status.Errorf(codes.Unknown, "cannot receive image info"))
	}
	imageName := req.GetInfo().GetName()
	imageType := req.GetInfo().ImageType
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
	imageName, err = s.filesMap.Save(imageName, imageType, imageData)
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

/*
func (s *server)ListImage(ctx context.Context,
	message *wrappers.StringValue)(*pb.ImageList, error){

	files, err:=ioutil.ReadDir("./temp")
	err!=nil{
		log.Fatal(err)
	}

	for _, f:=range files{
		stats, err:=os.Stat(f)
		log.Println("filename : ", stats.Name())
		log.Println("filename : ", stats.Size())
		log.Println("filename : ", stats.ModTime())
	}
}
*/

//func (s *server)DownloadImage(imageName *wrappers.StringValue,
//		stream pb.ImageUploadService_DownloadloadImageServer)error{

//check filename existence in map or directory
//send file by chunk
//}
