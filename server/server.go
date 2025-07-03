package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	pb "dist-kv/proto"
)

const storeFile = "store.txt"

// Server is used to implement KV service
type Server struct {
	pb.UnimplementedKVServer
	mu sync.RWMutex // Mutex for concurrent file access
}

// NewServer creates a new KV server instance
func NewServer() *Server {
	return &Server{}
}

// Put implements KV.Put
func (s *Server) Put(ctx context.Context, in *pb.PutRequest) (*pb.PutResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Open file in append mode
	file, err := os.OpenFile(storeFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &pb.PutResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to open store file: %v", err),
		}, nil
	}
	defer file.Close()

	// Write key=value to file
	entry := fmt.Sprintf("%s=%s\n", in.GetKey(), in.GetValue())
	if _, err := file.WriteString(entry); err != nil {
		return &pb.PutResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to write to store file: %v", err),
		}, nil
	}

	return &pb.PutResponse{
		Success: true,
		Message: "Key-value pair stored successfully",
	}, nil
}

// Get implements KV.Get
func (s *Server) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Open file for reading
	file, err := os.Open(storeFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &pb.GetResponse{
				Success: false,
				Value:   "",
				Message: "Key not found",
			}, nil
		}
		return &pb.GetResponse{
			Success: false,
			Value:   "",
			Message: fmt.Sprintf("Failed to open store file: %v", err),
		}, nil
	}
	defer file.Close()

	// Read file line by line to find the latest value for the key
	var latestValue string
	found := false
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 && parts[0] == in.GetKey() {
			latestValue = parts[1]
			found = true
		}
	}

	if err := scanner.Err(); err != nil {
		return &pb.GetResponse{
			Success: false,
			Value:   "",
			Message: fmt.Sprintf("Failed to read store file: %v", err),
		}, nil
	}

	if !found {
		return &pb.GetResponse{
			Success: false,
			Value:   "",
			Message: "Key not found",
		}, nil
	}

	return &pb.GetResponse{
		Success: true,
		Value:   latestValue,
		Message: "Key found",
	}, nil
}
