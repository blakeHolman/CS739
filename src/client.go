package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
	pb "madkv/kvstore"
)

// Global client instance
var client pb.KeyValueStoreClient

func put(key, value string) {
	resp, err := client.Put(context.Background(), &pb.PutRequest{Key: key, Value: value})
	if err != nil {
		fmt.Printf("PUT %s not_found, error\n", key)
		return
	}
	if resp.Found {
		fmt.Printf("PUT %s found\n", key)
	} else {
		fmt.Printf("PUT %s not_found\n", key)
	}
}

func get(key string) {
	resp, err := client.Get(context.Background(), &pb.GetRequest{Key: key})
	if err != nil || resp.Value == nil {
		fmt.Printf("GET %s null\n", key)
		return
	}
	fmt.Printf("GET %s %s\n", key, *resp.Value)
}

func swap(key, newValue string) {
	resp, err := client.Swap(context.Background(), &pb.SwapRequest{Key: key, NewValue: newValue})
	if err != nil || resp.OldValue == nil {
		fmt.Printf("SWAP %s null\n", key)
		return
	}
	fmt.Printf("SWAP %s %s\n", key, *resp.OldValue)
}

func deleteKey(key string) {
	resp, err := client.Delete(context.Background(), &pb.DeleteRequest{Key: key})
	if err != nil {
		fmt.Printf("DELETE %s not_found\n", key)
		return
	}
	if resp.Found {
		fmt.Printf("DELETE %s found\n", key)
	} else {
		fmt.Printf("DELETE %s not_found\n", key)
	}
}

func scan(startKey, endKey string) {
	resp, err := client.Scan(context.Background(), &pb.ScanRequest{StartKey: startKey, EndKey: endKey})
	if err != nil {
		fmt.Printf("SCAN %s %s ERROR\n", startKey, endKey)
		return
	}
	fmt.Printf("SCAN %s %s BEGIN\n", startKey, endKey)
	for i, key := range resp.Keys {
		if i < len(resp.Values) { 
			fmt.Printf("  %s %s\n", key, resp.Values[i])
		}
	}
	fmt.Println("SCAN END")
}

// Reads input and executes KV store commands
func processInput(interactive bool) {
	scanner := bufio.NewScanner(os.Stdin)

	if interactive {
		fmt.Println("KV Store Client Interactive Mode")
		fmt.Println("Enter commands (PUT key value, GET key, DELETE key, etc.). Type STOP to exit.")
		fmt.Print("> ")
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := strings.ToUpper(parts[0])

		switch command {
		case "PUT":
			if len(parts) < 3 {
				fmt.Println("Invalid PUT command. Usage: PUT <key> <value>")
				fmt.Print("> ")
				continue
			}
			put(parts[1], parts[2])
		case "GET":
			if len(parts) < 2 {
				fmt.Println("Invalid GET command. Usage: GET <key>")
				fmt.Print("> ")
				continue
			}
			get(parts[1])
		case "SWAP":
			if len(parts) < 3 {
				fmt.Println("Invalid SWAP command. Usage: SWAP <key> <new_value>")
				fmt.Print("> ")
				continue
			}
			swap(parts[1], parts[2])
		case "DELETE":
			if len(parts) < 2 {
				fmt.Println("Invalid DELETE command. Usage: DELETE <key>")
				fmt.Print("> ")
				continue
			}
			deleteKey(parts[1])
		case "SCAN":
			if len(parts) < 3 {
				fmt.Println("Invalid SCAN command. Usage: SCAN <start_key> <end_key>")
				fmt.Print("> ")
				continue
			}
			scan(parts[1], parts[2])
		case "STOP":
			fmt.Println("STOP")
			return
		default:
			fmt.Printf("Unknown command: %s\n", command)
		}

		// If interactive, prompt for the next command
		if interactive {
			fmt.Print("> ")
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading input: %v", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s [server_address]", os.Args[0])
	}
	serverAddress := os.Args[1]

	// Establish gRPC connection
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to server %s: %v", serverAddress, err)
	}
	defer conn.Close()

	client = pb.NewKeyValueStoreClient(conn)

	// Check if running interactively (if input is from a terminal)
	fi, err := os.Stdin.Stat()
	interactive := (err == nil && fi.Mode()&os.ModeCharDevice != 0)

	// Process input (interactive mode or batch mode)
	processInput(interactive)
}
