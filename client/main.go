package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	pb "github.com/AlexeyBurmak/image_service/gen/fileservice"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("gRPC connection error: %v", err.Error())
	}
	defer conn.Close()

	client := pb.NewFileServiceClient(conn)

	for i := 1; i <= 15; i++ {
		fileName := "logo_" + strconv.Itoa(i) + ".jpg"
		err = uploadFile(ctx, client, fileName)
		if err != nil {
			log.Fatalf("upload failed: %v", err)
		}
	}

	for i := 1; i <= 15; i++ {
		fileName := "logo_" + strconv.Itoa(i) + ".jpg"
		err = downloadFile(ctx, client, fileName)
		if err != nil {
			log.Fatalf("download failed: %v", err)
		}
	}

	err = listFiles(ctx, client)
	if err != nil {
		log.Fatalf("list files failed: %v", err)
	}
}

func uploadFile(ctx context.Context, client pb.FileServiceClient, path string) error {
	pathUpload := "./client/upload/"

	data, err := os.ReadFile(pathUpload + path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileMeta, err := os.Stat(pathUpload + path)
	if err != nil {
		return fmt.Errorf("failed to read file's metafata: %w", err)
	}
	createdAt := fileMeta.ModTime().Format(time.UnixDate)

	req := &pb.UploadRequest{
		Filename:  path,
		FileData:  data,
		CreatedAt: createdAt,
	}

	res, err := client.Upload(ctx, req)
	if err != nil {
		return fmt.Errorf("upload error: %w", err)
	}

	fmt.Println("Upload response:", res.GetMessage())
	return nil
}

func downloadFile(ctx context.Context, client pb.FileServiceClient, fileName string) error {
	pathDownload := "./client/download/"

	req := &pb.DownloadRequest{Filename: fileName}

	res, err := client.Download(ctx, req)
	if err != nil {
		return fmt.Errorf("download error: %w", err)
	}

	err = os.WriteFile(pathDownload+fileName, res.GetFileData(), 0644)
	if err != nil {
		return fmt.Errorf("write file error: %w", err)
	}

	fmt.Println("Downloaded: ", fileName)
	return nil
}

func listFiles(ctx context.Context, client pb.FileServiceClient) error {
	res, err := client.ListFiles(ctx, &pb.ListFilesRequest{})
	if err != nil {
		return fmt.Errorf("list error: %w", err)
	}

	fmt.Println("Files on server:")
	for _, file := range res.GetFiles() {
		fmt.Printf("- %s | created: %s | updated: %s\n",
			file.Filename, file.CreatedAt, file.UpdatedAt)
	}

	return nil
}
