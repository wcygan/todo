package main

// Updated with latest protobuf dependencies including UpdateTask
import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/grpcreflect"
	taskconnect "buf.build/gen/go/wcygan/todo/connectrpc/go/task/v1/taskv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/wcygan/todo/backend/internal/config"
	"github.com/wcygan/todo/backend/internal/handler"
	"github.com/wcygan/todo/backend/internal/logger"
	"github.com/wcygan/todo/backend/internal/middleware"
	"github.com/wcygan/todo/backend/internal/service"
	"github.com/wcygan/todo/backend/internal/store"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.New(cfg)
	log.LogInfo(context.Background(), "starting Todo ConnectRPC server", 
		"port", cfg.Server.Port,
		"development", cfg.IsDevelopment(),
		"log_level", cfg.Logger.Level,
	)

	// Initialize database store manager
	storeManager, err := store.NewManager(cfg)
	if err != nil {
		log.LogError(context.Background(), "failed to initialize store manager", err)
		os.Exit(1)
	}
	defer func() {
		if err := storeManager.Close(); err != nil {
			log.LogError(context.Background(), "failed to close store manager", err)
		}
	}()

	// Initialize dependencies with logging
	taskService := service.NewTaskService(storeManager.TaskStore())
	taskHandler := handler.NewTaskHandler(taskService)

	log.LogInfo(context.Background(), "dependencies initialized")

	// Create HTTP mux
	mux := http.NewServeMux()

	// Register health endpoint with database check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Check MySQL database health
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		
		if err := storeManager.HealthCheck(ctx); err != nil {
			log.LogError(ctx, "MySQL health check failed", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","service":"todo-backend","error":"mysql_unavailable"}`))
			return
		}
		
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"todo-backend","database":"mysql","store":"mysql"}`))
	})
	log.LogInfo(context.Background(), "health endpoint registered", "path", "/health")

	// Register TaskService
	path, serviceHandler := taskconnect.NewTaskServiceHandler(taskHandler)
	mux.Handle(path, serviceHandler)
	log.LogInfo(context.Background(), "task service registered", "path", path)

	// Add reflection support for development and testing
	reflector := grpcreflect.NewStaticReflector(
		taskconnect.TaskServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	log.LogInfo(context.Background(), "grpc reflection enabled")

	// Add CORS support for web clients
	corsHandler := createCORSHandler(mux, cfg, log)

	// Add timeout middleware
	timeoutHandler := middleware.TimeoutMiddleware(cfg, log)(corsHandler)

	// Add request logging middleware
	loggedHandler := logger.RequestLoggingMiddleware(log)(timeoutHandler)

	// Support HTTP/2 without TLS for local development
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      h2c.NewHandler(loggedHandler, &http2.Server{}),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.LogInfo(context.Background(), "server listening", 
			"addr", server.Addr,
			"endpoints", []string{
				"/health",
				path + "/CreateTask",
				path + "/GetAllTasks", 
				path + "/DeleteTask",
			},
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.LogError(context.Background(), "server failed to start", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.LogInfo(context.Background(), "shutting down server")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.LogError(context.Background(), "server forced to shutdown", err)
		os.Exit(1)
	}

	log.LogInfo(context.Background(), "server shutdown complete")
}

func createCORSHandler(mux *http.ServeMux, cfg *config.Config, log *logger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers based on configuration
		for _, origin := range cfg.Server.CORS.AllowedOrigins {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", 
			joinStrings(cfg.Server.CORS.AllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", 
			joinStrings(cfg.Server.CORS.AllowedHeaders, ", "))

		if r.Method == "OPTIONS" {
			log.LogDebug(r.Context(), "cors preflight request", "origin", r.Header.Get("Origin"))
			w.WriteHeader(http.StatusOK)
			return
		}

		mux.ServeHTTP(w, r)
	})
}

func joinStrings(slice []string, separator string) string {
	if len(slice) == 0 {
		return ""
	}
	
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += separator + slice[i]
	}
	return result
}