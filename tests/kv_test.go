package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	pb "dist-kv/proto"
)

const (
	bufSize   = 1024 * 1024
	storeFile = "store.txt"
)

var lis *bufconn.Listener

// TestServer implements the same logic as the main server for testing
type TestServer struct {
	pb.UnimplementedKVServer
	mu sync.RWMutex
}

// Put implements KV.Put
func (s *TestServer) Put(ctx context.Context, in *pb.PutRequest) (*pb.PutResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(storeFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &pb.PutResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to open store file: %v", err),
		}, nil
	}
	defer file.Close()

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
func (s *TestServer) Get(ctx context.Context, in *pb.GetRequest) (*pb.GetResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterKVServer(s, &TestServer{})
	
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(fmt.Sprintf("Server exited with error: %v", err))
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func getTestClient(t *testing.T) pb.KVClient {
	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return pb.NewKVClient(conn)
}

func TestPutAndGet(t *testing.T) {
	// Clean up store file before test
	os.Remove("store.txt")
	defer os.Remove("store.txt")

	client := getTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test Put
	putResp, err := client.Put(ctx, &pb.PutRequest{
		Key:   "test_key",
		Value: "test_value",
	})
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	if !putResp.Success {
		t.Fatalf("Put was not successful: %s", putResp.Message)
	}

	// Test Get
	getResp, err := client.Get(ctx, &pb.GetRequest{
		Key: "test_key",
	})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !getResp.Success {
		t.Fatalf("Get was not successful: %s", getResp.Message)
	}
	if getResp.Value != "test_value" {
		t.Fatalf("Expected value 'test_value', got '%s'", getResp.Value)
	}
}

func TestGetNonExistentKey(t *testing.T) {
	// Clean up store file before test
	os.Remove("store.txt")
	defer os.Remove("store.txt")

	client := getTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test Get non-existent key
	getResp, err := client.Get(ctx, &pb.GetRequest{
		Key: "non_existent_key",
	})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if getResp.Success {
		t.Fatalf("Get should have failed for non-existent key")
	}
	if getResp.Value != "" {
		t.Fatalf("Expected empty value for non-existent key, got '%s'", getResp.Value)
	}
}

func TestMultiplePutsWithSameKey(t *testing.T) {
	// Clean up store file before test
	os.Remove("store.txt")
	defer os.Remove("store.txt")

	client := getTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Put first value
	putResp1, err := client.Put(ctx, &pb.PutRequest{
		Key:   "same_key",
		Value: "first_value",
	})
	if err != nil {
		t.Fatalf("First Put failed: %v", err)
	}
	if !putResp1.Success {
		t.Fatalf("First Put was not successful: %s", putResp1.Message)
	}

	// Put second value with same key
	putResp2, err := client.Put(ctx, &pb.PutRequest{
		Key:   "same_key",
		Value: "second_value",
	})
	if err != nil {
		t.Fatalf("Second Put failed: %v", err)
	}
	if !putResp2.Success {
		t.Fatalf("Second Put was not successful: %s", putResp2.Message)
	}

	// Get should return the latest value
	getResp, err := client.Get(ctx, &pb.GetRequest{
		Key: "same_key",
	})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if !getResp.Success {
		t.Fatalf("Get was not successful: %s", getResp.Message)
	}
	if getResp.Value != "second_value" {
		t.Fatalf("Expected latest value 'second_value', got '%s'", getResp.Value)
	}
}

func TestMultipleKeys(t *testing.T) {
	// Clean up store file before test
	os.Remove("store.txt")
	defer os.Remove("store.txt")

	client := getTestClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Put multiple key-value pairs
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for key, value := range testData {
		putResp, err := client.Put(ctx, &pb.PutRequest{
			Key:   key,
			Value: value,
		})
		if err != nil {
			t.Fatalf("Put failed for key %s: %v", key, err)
		}
		if !putResp.Success {
			t.Fatalf("Put was not successful for key %s: %s", key, putResp.Message)
		}
	}

	// Get all values and verify
	for key, expectedValue := range testData {
		getResp, err := client.Get(ctx, &pb.GetRequest{
			Key: key,
		})
		if err != nil {
			t.Fatalf("Get failed for key %s: %v", key, err)
		}
		if !getResp.Success {
			t.Fatalf("Get was not successful for key %s: %s", key, getResp.Message)
		}
		if getResp.Value != expectedValue {
			t.Fatalf("Expected value '%s' for key '%s', got '%s'", expectedValue, key, getResp.Value)
		}
	}
}
