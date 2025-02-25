package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	//_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/proto"

	pb "madkv/kvstore" // Import generated Go package

	"google.golang.org/grpc"
)

type KeyValueStoreServer struct {
	pb.UnimplementedKeyValueStoreServer
	mu    sync.RWMutex
	store map[string]string
	db    *sql.DB
}

// Initialize DB
func initDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS command_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sommand_type TEXT NOT NULL,
		serialized_data BLOB NOT NULL,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	return db, nil

}

// Logging Function
func (s *KeyValueStoreServer) logCommand(commandType string, msg proto.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = s.db.Exec("INSERT INTO command_log (command_type, serialized_data) VALUES (?, ?)", commandType, data)
	return err
}

// Replay log during startup ###### UNFINISHED ######
func (s *KeyValueStoreServer) replayLog() error {
	rows, err := s.db.Query("SELECT command_type, serialized_data FROM command_log ORDER BY id ASC")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var commandType string
		var serializedData []byte
		if err := rows.Scan(&commandType, &serializedData); err != nil {
			return err
		}

		switch commandType {
		case "PUT":
			var req pb.PutRequest
			if err := proto.Unmarshal(serializedData, &req); err != nil {
				return err
			}
			s.store[req.Key] = req.Value
		case "DELETE":
			var req pb.DeleteRequest
			if err := proto.Unmarshal(serializedData, &req); err != nil {
				return err
			}
			delete(s.store, req.Key)
		case "SWAP":
			var req pb.SwapRequest
			if err := proto.Unmarshal(serializedData, &req); err != nil {
				return err
			}

			_, found := s.store[req.Key]
			if found {
				s.store[req.Key] = req.NewValue
			}
		}
	}

	// Catch any errors during iteration
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// PUT: Insert or update a key-value pair
func (s *KeyValueStoreServer) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	s.mu.Lock() // Write lock
	defer s.mu.Unlock()

	// Log the command first
	err := s.logCommand("PUT", req)
	if err != nil {
		return nil, err
	}

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

func main() { // Cannot have two func mains in the same package (server.go and client.go)

	db, err := initDB("command_log.db")
	if err != nil {
		log.Fatalf("Failed to initialize DB: %v", err)
	}

	server := grpc.NewServer()
	kvServer := &KeyValueStoreServer{
		store: make(map[string]string),
		db:    db,
	}

	// Replay the logs
	if err := kvServer.replayLog(); err != nil {
		log.Fatalf("Failed to replay logs: %v", err)
	}

	listenAddr := flag.String("listen", "0.0.0.0:3777", "IP:Port address to listen on")
	flag.Parse()

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", *listenAddr, err)
	}

	//server := grpc.NewServer()
	pb.RegisterKeyValueStoreServer(server, &KeyValueStoreServer{store: make(map[string]string)})

	log.Printf("Server listening on %s", *listenAddr)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
