package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "dist-kv/proto"
)

const (
	address = "localhost:50051"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run client/main.go put <key> <value>")
		fmt.Println("  go run client/main.go get <key>")
		os.Exit(1)
	}

	// Set up a connection to the server
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewKVClient(conn)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	command := os.Args[1]

	switch command {
	case "put":
		if len(os.Args) != 4 {
			fmt.Println("Usage: go run client/main.go put <key> <value>")
			os.Exit(1)
		}
		key := os.Args[2]
		value := os.Args[3]
		
		putClient(ctx, client, key, value)

	case "get":
		if len(os.Args) != 3 {
			fmt.Println("Usage: go run client/main.go get <key>")
			os.Exit(1)
		}
		key := os.Args[2]
		
		getClient(ctx, client, key)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: put, get")
		os.Exit(1)
	}
}

func putClient(ctx context.Context, client pb.KVClient, key, value string) {
	fmt.Printf("Storing key='%s' value='%s'\n", key, value)
	
	response, err := client.Put(ctx, &pb.PutRequest{
		Key:   key,
		Value: value,
	})
	
	if err != nil {
		log.Fatalf("Put failed: %v", err)
	}
	
	if response.Success {
		fmt.Printf("✓ Success: %s\n", response.Message)
	} else {
		fmt.Printf("✗ Failed: %s\n", response.Message)
	}
}

func getClient(ctx context.Context, client pb.KVClient, key string) {
	fmt.Printf("Retrieving value for key='%s'\n", key)
	
	response, err := client.Get(ctx, &pb.GetRequest{
		Key: key,
	})
	
	if err != nil {
		log.Fatalf("Get failed: %v", err)
	}
	
	if response.Success {
		fmt.Printf("✓ Found: key='%s' value='%s'\n", key, response.Value)
	} else {
		fmt.Printf("✗ Not found: %s\n", response.Message)
	}
}
