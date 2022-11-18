# File upload and download with GRPC (implemented with Golang)

Сервис :
cd server
go run *.go

Клиент :
cd client
go run main.go

- Чтобы получить списка файлов:
 getImagesList(c)
-  Скачать файла от сервиса :
 DownloadImage(c, название файла)
- Загрузить файла:
  uploadImage(c, папка/назавание файла)
