package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	taskconnect "buf.build/gen/go/wcygan/todo/connectrpc/go/task/v1/taskv1connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/wcygan/todo/backend/internal/handler"
	"github.com/wcygan/todo/backend/internal/service"
	"github.com/wcygan/todo/backend/test/testutil"
)

// setupTestServer creates a test server with the full application stack
func setupTestServer() (*httptest.Server, taskconnect.TaskServiceClient) {
	// Create dependencies
	taskStore := testutil.NewMockStore()
	taskService := service.NewTaskService(taskStore)
	taskHandler := handler.NewTaskHandler(taskService)
	
	// Create HTTP mux
	mux := http.NewServeMux()
	
	// Register TaskService
	path, serviceHandler := taskconnect.NewTaskServiceHandler(taskHandler)
	mux.Handle(path, serviceHandler)
	
	// Add reflection support
	reflector := grpcreflect.NewStaticReflector(
		taskconnect.TaskServiceName,
	)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	
	// Add CORS support
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
	
	// Create test server with HTTP/2 support
	server := httptest.NewUnstartedServer(
		h2c.NewHandler(http.HandlerFunc(corsHandler), &http2.Server{}),
	)
	server.EnableHTTP2 = true
	server.Start()
	
	// Create client
	client := taskconnect.NewTaskServiceClient(
		http.DefaultClient,
		server.URL,
	)
	
	return server, client
}

func TestIntegration_FullTaskWorkflow(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()
	
	ctx := context.Background()
	
	// 1. Initially, no tasks should exist
	getAllResp, err := client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)
	assert.Empty(t, getAllResp.Msg.Tasks, "Initially should have no tasks")
	
	// 2. Create first task
	createResp1, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "First integration test task",
	}))
	require.NoError(t, err)
	require.NotNil(t, createResp1.Msg.Task)
	
	task1 := createResp1.Msg.Task
	assert.Equal(t, "1", task1.Id)
	assert.Equal(t, "First integration test task", task1.Description)
	assert.False(t, task1.Completed)
	assert.NotNil(t, task1.CreatedAt)
	assert.NotNil(t, task1.UpdatedAt)
	
	// 3. Create second task
	createResp2, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "Second integration test task",
	}))
	require.NoError(t, err)
	require.NotNil(t, createResp2.Msg.Task)
	
	task2 := createResp2.Msg.Task
	assert.Equal(t, "2", task2.Id)
	assert.Equal(t, "Second integration test task", task2.Description)
	
	// 4. Get all tasks - should have 2
	getAllResp, err = client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)
	require.Len(t, getAllResp.Msg.Tasks, 2, "Should have 2 tasks after creating 2")
	
	// Verify both tasks are present
	taskMap := make(map[string]*taskv1.Task)
	for _, task := range getAllResp.Msg.Tasks {
		taskMap[task.Id] = task
	}
	
	assert.Contains(t, taskMap, task1.Id, "Task 1 should be in the list")
	assert.Contains(t, taskMap, task2.Id, "Task 2 should be in the list")
	assert.Equal(t, task1.Description, taskMap[task1.Id].Description)
	assert.Equal(t, task2.Description, taskMap[task2.Id].Description)
	
	// 5. Delete first task
	deleteResp, err := client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{
		Id: task1.Id,
	}))
	require.NoError(t, err)
	assert.True(t, deleteResp.Msg.Success, "Deletion should be successful")
	assert.Equal(t, "Task deleted successfully", deleteResp.Msg.Message)
	
	// 6. Get all tasks - should have 1
	getAllResp, err = client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)
	require.Len(t, getAllResp.Msg.Tasks, 1, "Should have 1 task after deleting 1")
	
	remainingTask := getAllResp.Msg.Tasks[0]
	assert.Equal(t, task2.Id, remainingTask.Id, "Remaining task should be task 2")
	assert.Equal(t, task2.Description, remainingTask.Description)
	
	// 7. Try to delete non-existent task
	deleteResp, err = client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{
		Id: "999",
	}))
	require.NoError(t, err)
	assert.False(t, deleteResp.Msg.Success, "Deletion of non-existent task should fail")
	assert.Contains(t, deleteResp.Msg.Message, "not found")
	
	// 8. Verify task count unchanged
	getAllResp, err = client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)
	assert.Len(t, getAllResp.Msg.Tasks, 1, "Task count should remain 1")
}

func TestIntegration_ConcurrentOperations(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()
	
	ctx := context.Background()
	
	const numGoroutines = 10
	const tasksPerGoroutine = 5
	
	// Create tasks concurrently
	done := make(chan string, numGoroutines*tasksPerGoroutine)
	
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < tasksPerGoroutine; j++ {
				createResp, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
					Description: "Concurrent task",
				}))
				assert.NoError(t, err)
				if err == nil && createResp.Msg.Task != nil {
					done <- createResp.Msg.Task.Id
				} else {
					done <- ""
				}
			}
		}(i)
	}
	
	// Collect all created task IDs
	createdIDs := make(map[string]bool)
	for i := 0; i < numGoroutines*tasksPerGoroutine; i++ {
		id := <-done
		if id != "" {
			assert.False(t, createdIDs[id], "No duplicate IDs should be created: %s", id)
			createdIDs[id] = true
		}
	}
	
	// Verify all tasks were created
	getAllResp, err := client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)
	assert.Len(t, getAllResp.Msg.Tasks, numGoroutines*tasksPerGoroutine, 
		"All concurrent tasks should be created")
	
	// Verify all created IDs are present in the final list
	finalTaskIDs := make(map[string]bool)
	for _, task := range getAllResp.Msg.Tasks {
		finalTaskIDs[task.Id] = true
	}
	
	for createdID := range createdIDs {
		assert.True(t, finalTaskIDs[createdID], 
			"Created task ID %s should be in final list", createdID)
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()
	
	ctx := context.Background()
	
	// Test deleting non-existent task
	deleteResp, err := client.DeleteTask(ctx, connect.NewRequest(&taskv1.DeleteTaskRequest{
		Id: "non-existent-id",
	}))
	
	require.NoError(t, err, "Should not return connection error")
	assert.False(t, deleteResp.Msg.Success, "Should indicate failure")
	assert.Contains(t, deleteResp.Msg.Message, "not found", "Should indicate task not found")
}

func TestIntegration_EmptyDescriptions(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()
	
	ctx := context.Background()
	
	// Create task with empty description should fail with validation error
	_, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "",
	}))
	
	require.Error(t, err, "Empty description should not be allowed")
	
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
	assert.Contains(t, connectErr.Message(), "description cannot be empty")
	
	// Verify no tasks were created
	getAllResp, err := client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)
	assert.Empty(t, getAllResp.Msg.Tasks, "No tasks should have been created")
}

func TestIntegration_LongDescriptions(t *testing.T) {
	server, client := setupTestServer()
	defer server.Close()
	
	ctx := context.Background()
	
	// Create task with very long description
	longDescription := "This is a very long task description that contains a lot of text to test how the system handles longer inputs. " +
		"It should be able to handle this without any issues and store the complete description properly. " +
		"The system should maintain data integrity regardless of the description length within reasonable bounds."
	
	createResp, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: longDescription,
	}))
	
	require.NoError(t, err)
	require.NotNil(t, createResp.Msg.Task)
	
	task := createResp.Msg.Task
	assert.Equal(t, longDescription, task.Description, "Long description should be preserved exactly")
	
	// Verify it appears correctly in the list
	getAllResp, err := client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
	require.NoError(t, err)
	require.Len(t, getAllResp.Msg.Tasks, 1)
	assert.Equal(t, longDescription, getAllResp.Msg.Tasks[0].Description)
}

func BenchmarkIntegration_CreateAndListTasks(b *testing.B) {
	server, client := setupTestServer()
	defer server.Close()
	
	ctx := context.Background()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Create a task
		_, err := client.CreateTask(ctx, connect.NewRequest(&taskv1.CreateTaskRequest{
			Description: "Benchmark task",
		}))
		if err != nil {
			b.Fatalf("Failed to create task: %v", err)
		}
		
		// List all tasks
		_, err = client.GetAllTasks(ctx, connect.NewRequest(&taskv1.GetAllTasksRequest{}))
		if err != nil {
			b.Fatalf("Failed to get tasks: %v", err)
		}
	}
}