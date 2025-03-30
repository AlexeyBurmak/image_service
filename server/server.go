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

func NewFileServiceServer(ctx context.Context, storagePath string, redisClient *redis.Client) pb.FileServiceServer {
	maxUploadDownload := 10
	maxList := 100

	server := &FileServiceServer{
		storagePath:   storagePath,
		redisClient:   redisClient,
		uploadQueue:   make(chan *UploadTask, maxUploadDownload),
		downloadQueue: make(chan *DownloadTask, maxUploadDownload),
		listQueue:     make(chan *ListFilesTask, maxList),
	}

	for i := 0; i < maxUploadDownload; i++ {
		go server.uploadWorker(ctx)
		go server.downloadWorker(ctx)
	}

	for i := 0; i < maxList; i++ {
		go server.listWorker(ctx)
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
		Resp:    make(chan *pb.ListFilesResponse, 1),
		ErrChan: make(chan error, 1),
	}

	select {
	case s.listQueue <- task:
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "request timeout")
	}

	select {
	case res := <-task.Resp:
		return res, nil
	case err := <-task.ErrChan:
		return nil, err
	case <-ctx.Done():
		return nil, status.Error(codes.DeadlineExceeded, "request timeout")
	}
}
