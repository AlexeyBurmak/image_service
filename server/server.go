package server

import (
	"context"

	pb "github.com/AlexeyBurmak/image_service/gen/fileservice"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FileServiceServer struct {
	pb.UnimplementedFileServiceServer
	storagePath   string
	redisClient   *redis.Client
	uploadQueue   chan *UploadTask
	downloadQueue chan *DownloadTask
	listQueue     chan *ListFilesTask
}

func NewFileServiceServer(storagePath string, redisClient *redis.Client) pb.FileServiceServer {
	server := &FileServiceServer{
		storagePath:   storagePath,
		redisClient:   redisClient,
		uploadQueue:   make(chan *UploadTask, 10),
		downloadQueue: make(chan *DownloadTask, 10),
		listQueue:     make(chan *ListFilesTask, 100),
	}

	for i := 0; i < 10; i++ {
		go server.uploadWorker()
	}

	for i := 0; i < 10; i++ {
		go server.downloadWorker()
	}

	for i := 0; i < 100; i++ {
		go server.listWorker()
	}

	return server
}

func (s *FileServiceServer) Upload(ctx context.Context, req *pb.UploadRequest) (*pb.UploadResponse, error) {
	task := &UploadTask{
		Req:  req,
		Done: make(chan uploadResult, 1),
	}

	select {
	case s.uploadQueue <- task:
		// accepted
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "upload timeout")
	}

	select {
	case result := <-task.Done:
		return result.Resp, result.Err
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "upload cancelled")
	}
}

func (s *FileServiceServer) Download(ctx context.Context, req *pb.DownloadRequest) (*pb.DownloadResponse, error) {
	task := &DownloadTask{
		Req:  req,
		Ctx:  ctx,
		Done: make(chan downloadResult, 1),
	}

	select {
	case s.downloadQueue <- task:
		// accepted
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "download timeout")
	}

	select {
	case result := <-task.Done:
		return result.Resp, result.Err
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "client cancelled")
	}
}

func (s *FileServiceServer) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	task := &ListFilesTask{
		Ctx:     ctx,
		Req:     req,
		Result:  make(chan *pb.ListFilesResponse, 1),
		ErrChan: make(chan error, 1),
	}

	select {
	case s.listQueue <- task:
		// Task accepted
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "request timeout")
	}

	select {
	case res := <-task.Result:
		return res, nil
	case err := <-task.ErrChan:
		return nil, err
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "request timeout")
	}
}

// var _ pb.FileServiceServer = (*FileServiceServer)(nil)
