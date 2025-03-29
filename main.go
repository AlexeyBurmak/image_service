package main

import (
	"log"
	"net"

	pb "github.com/AlexeyBurmak/image_service/gen/fileservice"
	"github.com/AlexeyBurmak/image_service/server"
	"github.com/redis/go-redis/v9"

	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // или имя контейнера, если в Docker
		DB:   0,
	})

	grpcServer := grpc.NewServer()
	srv := server.NewFileServiceServer("storage/files", rdb)
	pb.RegisterFileServiceServer(grpcServer, srv)

	log.Println("gRPC server is running on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
