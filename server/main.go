package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
	pb "dist-kv/proto"
)

const port = ":50051"

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterKVServer(s, NewServer())

	log.Printf("gRPC server listening at %v", lis.Addr())
	log.Printf("Store file: store.txt")
	
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
