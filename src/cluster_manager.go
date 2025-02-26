package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"sync"

	pb "madkv/kvstore" // Import generated gRPC package

	"google.golang.org/grpc"
)

// ClusterManager struct
type ClusterManager struct {
	mu           sync.RWMutex
	serverList   map[int]string // Maps server_id -> address
	partitionMap map[string]int // Maps key -> server_id
}

// Hash function for consistent key partitioning
func hash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32())
}

// RegisterServer (gRPC function)
func (cm *ClusterManager) RegisterServer(ctx context.Context, req *pb.RegisterServerRequest) (*pb.RegisterServerResponse, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	server_id := int(req.ServerId)
	address := req.Address

	if _, found := cm.serverList[server_id]; found {
		return nil, fmt.Errorf("[ERROR] Server %d already registered", server_id)
	}

	cm.serverList[server_id] = address
	fmt.Printf("[REGISTER] Server %d registered at %s\n", server_id, address)

	return &pb.RegisterServerResponse{Success: true}, nil
}

// Assign partitions to servers (hash-based)
func (cm *ClusterManager) assignPartition(key string) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	num_servers := len(cm.serverList)
	if num_servers == 0 {
		return -1 // No servers available
	}

	assigned_server := hash(key) % num_servers
	return assigned_server
}

// Get partition info for a server
func (cm *ClusterManager) GetPartitionInfo(server_id int) ([]string, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if _, found := cm.serverList[server_id]; !found {
		return nil, fmt.Errorf("[ERROR] Server %d is not registered", server_id)
	}

	// Find all keys assigned to this server
	var keys []string
	for key, assignedServer := range cm.partitionMap {
		if assignedServer == server_id {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// GetClusterInfo (Returns all partitions)
func (cm *ClusterManager) GetClusterInfo() map[string]int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.partitionMap
}

// Main function to start the Cluster Manager
func main() {
	// Initialize ClusterManager
	cm := &ClusterManager{
		serverList:   make(map[int]string),
		partitionMap: make(map[string]int),
	}

	// Start gRPC server
	listenAddr := ":50051"
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", listenAddr, err)
	}

	server := grpc.NewServer()
	pb.RegisterClusterManagerServer(server, cm)

	log.Printf("Cluster Manager started on port %s", listenAddr)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
