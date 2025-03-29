package main

import (
	"context"
	"fmt"
	"log"
	"os"
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

	// Загрузка файла
	err = uploadFile(ctx, client, "test_upload.txt")
	if err != nil {
		log.Fatalf("upload failed: %v", err)
	}

	// Скачивание файла
	err = downloadFile(ctx, client, "test_upload.txt", "downloaded.txt")
	if err != nil {
		log.Fatalf("download failed: %v", err)
	}

	// Получение списка файлов
	err = listFiles(ctx, client)
	if err != nil {
		log.Fatalf("list files failed: %v", err)
	}
}

func uploadFile(ctx context.Context, client pb.FileServiceClient, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	fmt.Println("Got the file!")

	req := &pb.UploadRequest{
		Filename: path,
		FileData: data,
	}

	res, err := client.Upload(ctx, req)
	if err != nil {
		return fmt.Errorf("upload error: %w", err)
	}
	fmt.Println("File Uploaded!")

	fmt.Println("Upload response:", res.GetMessage())
	return nil
}

func downloadFile(ctx context.Context, client pb.FileServiceClient, remoteName, localName string) error {
	req := &pb.DownloadRequest{Filename: remoteName}

	res, err := client.Download(ctx, req)
	if err != nil {
		return fmt.Errorf("download error: %w", err)
	}

	err = os.WriteFile(localName, res.GetFileData(), 0644)
	if err != nil {
		return fmt.Errorf("write file error: %w", err)
	}

	fmt.Println("Downloaded and saved as:", localName)
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
