package main

import (
	"fmt"
	"log"
	"net/http"

	"connectrpc.com/grpcreflect"
	taskconnect "buf.build/gen/go/wcygan/todo/connectrpc/go/task/v1/taskv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/wcygan/todo/backend/internal/handler"
	"github.com/wcygan/todo/backend/internal/store"
)

func main() {
	// Initialize dependencies
	taskStore := store.New()
	taskService := handler.NewTaskService(taskStore)

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register TaskService
	path, serviceHandler := taskconnect.NewTaskServiceHandler(taskService)
	mux.Handle(path, serviceHandler)

	// Add reflection support for development and testing
	reflector := grpcreflect.NewStaticReflector(
		taskconnect.TaskServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Add CORS support for web clients
	corsHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Connect-Protocol-Version, Connect-Timeout-Ms")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		mux.ServeHTTP(w, r)
	}

	// Support HTTP/2 without TLS for local development
	server := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(http.HandlerFunc(corsHandler), &http2.Server{}),
	}

	fmt.Println("Todo ConnectRPC server starting on :8080")
	fmt.Println("Reflection enabled for grpcurl/Postman testing")
	fmt.Println("Available endpoints:")
	fmt.Printf("  - %s/CreateTask\n", path)
	fmt.Printf("  - %s/GetAllTasks\n", path)
	fmt.Printf("  - %s/DeleteTask\n", path)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}