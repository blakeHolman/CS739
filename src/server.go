package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"flag"

	pb "madkv/kvstore" // Import generated Go package

	"google.golang.org/grpc"
)

type KeyValueStoreServer struct {
	pb.UnimplementedKeyValueStoreServer
	mu    sync.RWMutex
	store map[string]string
}

// PUT: Insert or update a key-value pair
func (s *KeyValueStoreServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	s.mu.Lock() // Write lock
	defer s.mu.Unlock()

	_, found := s.store[req.Key]
	s.store[req.Key] = req.Value
	fmt.Printf("[PUT] %s -> %s\n", req.Key, req.Value)

	return &pb.PutResponse{Found: found}, nil
}

// GET: Retrieve a value by key
func (s *KeyValueStoreServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	s.mu.RLock() // Read lock
	defer s.mu.RUnlock()

	value, found := s.store[req.Key]
	if !found {
		fmt.Printf("[GET] %s not found\n", req.Key)
		return &pb.GetResponse{}, nil
	}

	fmt.Printf("[GET] %s -> %s\n", req.Key, value)
	return &pb.GetResponse{Value: &value}, nil
}

// SWAP: Update a key's value and return the old value
func (s *KeyValueStoreServer) Swap(ctx context.Context, req *pb.SwapRequest) (*pb.SwapResponse, error) {
	s.mu.Lock() // Write lock
	defer s.mu.Unlock()

	oldValue, found := s.store[req.Key]
	s.store[req.Key] = req.NewValue
	if found {
		fmt.Printf("[SWAP] %s swapped %s -> %s\n", req.Key, oldValue, req.NewValue)
		return &pb.SwapResponse{OldValue: &oldValue}, nil
	}

	fmt.Printf("[SWAP] %s not found\n", req.Key)
	return &pb.SwapResponse{}, nil
}

// DELETE: Remove a key-value pair
func (s *KeyValueStoreServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	s.mu.Lock() // Write lock
	defer s.mu.Unlock()

	_, found := s.store[req.Key]
	if found {
		delete(s.store, req.Key)
		fmt.Printf("[DELETE] %s removed\n", req.Key)
		return &pb.DeleteResponse{Found: true}, nil
	}

	fmt.Printf("[DELETE] %s not found\n", req.Key)
	return &pb.DeleteResponse{Found: false}, nil
}

// SCAN: Retrieve all key-value pairs in a given range
func (s *KeyValueStoreServer) Scan(ctx context.Context, req *pb.ScanRequest) (*pb.ScanResponse, error) {
	s.mu.RLock() // Read lock
	defer s.mu.RUnlock()

	response := &pb.ScanResponse{}
	for key, value := range s.store {
		if key >= req.StartKey && key <= req.EndKey {
			response.Keys = append(response.Keys, key)
			response.Values = append(response.Values, value)
		}
	}

	fmt.Printf("[SCAN] %s to %s -> %d keys found\n", req.StartKey, req.EndKey, len(response.Keys))
	return response, nil
}

func main() {
	listenAddr := flag.String("listen", "0.0.0.0:3777", "IP:Port address to listen on")
	flag.Parse()

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", *listenAddr, err)
	}

	server := grpc.NewServer()
	pb.RegisterKeyValueStoreServer(server, &KeyValueStoreServer{store: make(map[string]string)})

	log.Printf("Server listening on %s", *listenAddr)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}