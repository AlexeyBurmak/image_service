package server

import (
	"context"

	"os"
	"path/filepath"
	"strings"
	"time"

	pb "github.com/AlexeyBurmak/image_service/gen/fileservice"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UploadTask struct {
	Req  *pb.UploadRequest
	Done chan uploadResult
}

type uploadResult struct {
	Resp *pb.UploadResponse
	Err  error
}

type DownloadTask struct {
	Req  *pb.DownloadRequest
	Ctx  context.Context
	Done chan downloadResult
}

type downloadResult struct {
	Resp *pb.DownloadResponse
	Err  error
}

type ListFilesTask struct {
	Ctx     context.Context
	Req     *pb.ListFilesRequest
	Resp    chan *pb.ListFilesResponse
	ErrChan chan error
}

func (s *FileServiceServer) uploadWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-s.uploadQueue:
			resp, err := s.performUpload(task.Req)
			task.Done <- uploadResult{Resp: resp, Err: err}
		}
	}
}

func (s *FileServiceServer) downloadWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-s.downloadQueue:
			resp, err := s.performDownload(task.Ctx, task.Req)
			task.Done <- downloadResult{Resp: resp, Err: err}
		}
	}

}

func (s *FileServiceServer) listWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-s.listQueue:
			res, err := s.performListFiles(ctx)
			if err != nil {
				task.ErrChan <- err
			} else {
				task.Resp <- res
			}
		}
	}
}

func (s *FileServiceServer) performUpload(req *pb.UploadRequest) (*pb.UploadResponse, error) {
	if req.GetFilename() == "" || len(req.GetFileData()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "non-empty filename and file data are required")
	}

	filePath := filepath.Join(s.storagePath, req.GetFilename())
	if err := os.WriteFile(filePath, req.GetFileData(), 0644); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save file: %v", err)
	}

	now := time.Now().UTC().Format(time.UnixDate)
	key := "filemeta:" + req.GetFilename()

	_, err := s.redisClient.HSet(context.Background(), key, map[string]interface{}{
		"created_at": req.CreatedAt,
		"updated_at": now,
	}).Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to write metadata to redis: %v", err)
	}

	return &pb.UploadResponse{
		Message: "Uploaded: " + req.Filename,
	}, nil
}

func (s *FileServiceServer) performDownload(ctx context.Context, req *pb.DownloadRequest) (*pb.DownloadResponse, error) {
	if req.Filename == "" {
		return nil, status.Error(codes.InvalidArgument, "filename required")
	}

	filePath := filepath.Join(s.storagePath, req.Filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, status.Error(codes.NotFound, "file not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to read file: %v", err)
	}

	return &pb.DownloadResponse{
		FileData: data,
	}, nil
}

func (s *FileServiceServer) performListFiles(ctx context.Context) (*pb.ListFilesResponse, error) {
	keys, err := s.redisClient.Keys(ctx, "filemeta:*").Result()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "redis error: %v", err)
	}

	var files []*pb.FileInfo

	for _, key := range keys {
		meta, err := s.redisClient.HGetAll(ctx, key).Result()
		if err != nil || len(meta) == 0 {
			continue
		}

		filename := strings.TrimPrefix(key, "filemeta:")
		files = append(files, &pb.FileInfo{
			Filename:  filename,
			CreatedAt: meta["created_at"],
			UpdatedAt: meta["updated_at"],
		})
	}

	return &pb.ListFilesResponse{Files: files}, nil
}
