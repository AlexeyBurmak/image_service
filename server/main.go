package main

import (
	"log"
	"net"

	pb "github.com/AlexeyBurmak/image_service/gen/fileservice"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // или "redis:6379", если будет docker-compose
		DB:   0,
	})

	grpcServer := grpc.NewServer()
	srv := NewFileServiceServer("storage/files", rdb)
	pb.RegisterFileServiceServer(grpcServer, srv)

	log.Println("gRPC server is running on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
