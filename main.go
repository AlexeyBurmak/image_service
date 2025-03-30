package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	pb "github.com/AlexeyBurmak/image_service/gen/fileservice"
	"github.com/AlexeyBurmak/image_service/server"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	port := ":50051"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	grpcServer := grpc.NewServer()
	srv := server.NewFileServiceServer(ctx, "storage", rdb)
	pb.RegisterFileServiceServer(grpcServer, srv)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully")
		time.Sleep(2 * time.Second)

		// =========== Redis db clean is optional ========
		err := rdb.FlushDB(ctx).Err()
		if err != nil {
			log.Printf("Redis cleanup error: %v", err)
		} else {
			log.Println("Redis cleaned")
		}

		cancel()
		grpcServer.GracefulStop()
	}()

	log.Printf("Image server is running on port %s\n", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to run service: %v", err)
	}

}
