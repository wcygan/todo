package handler

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wcygan/todo/backend/internal/store"
)

func TestNewTaskService(t *testing.T) {
	taskStore := store.New()
	service := NewTaskService(taskStore)
	
	assert.NotNil(t, service)
	assert.Equal(t, taskStore, service.store)
}

func TestTaskService_CreateTask(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantErr     bool
	}{
		{
			name:        "create_valid_task",
			description: "Test task description",
			wantErr:     false,
		},
		{
			name:        "create_empty_description",
			description: "",
			wantErr:     false,
		},
		{
			name:        "create_long_description",
			description: "This is a very long task description that should be handled properly by the system",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			taskStore := store.New()
			service := NewTaskService(taskStore)
			ctx := context.Background()
			
			req := connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: tt.description,
			})
			
			// Execute
			resp, err := service.CreateTask(ctx, req)
			
			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotNil(t, resp.Msg)
				require.NotNil(t, resp.Msg.Task)
				
				task := resp.Msg.Task
				assert.Equal(t, "1", task.Id)
				assert.Equal(t, tt.description, task.Description)
				assert.False(t, task.Completed)
				assert.NotNil(t, task.CreatedAt)
				assert.NotNil(t, task.UpdatedAt)
			}
		})
	}
}

func TestTaskService_GetAllTasks(t *testing.T) {
	tests := []struct {
		name          string
		setupTasks    []string
		expectedCount int
	}{
		{
			name:          "get_empty_list",
			setupTasks:    []string{},
			expectedCount: 0,
		},
		{
			name:          "get_single_task",
			setupTasks:    []string{"Task 1"},
			expectedCount: 1,
		},
		{
			name:          "get_multiple_tasks",
			setupTasks:    []string{"Task 1", "Task 2", "Task 3"},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			taskStore := store.New()
			service := NewTaskService(taskStore)
			ctx := context.Background()
			
			// Create setup tasks
			for _, desc := range tt.setupTasks {
				taskStore.CreateTask(desc)
			}
			
			req := connect.NewRequest(&taskv1.GetAllTasksRequest{})
			
			// Execute
			resp, err := service.GetAllTasks(ctx, req)
			
			// Assert
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			
			tasks := resp.Msg.Tasks
			assert.Len(t, tasks, tt.expectedCount)
			
			// Verify task contents if any exist
			if len(tt.setupTasks) > 0 {
				taskDescriptions := make(map[string]bool)
				for _, task := range tasks {
					taskDescriptions[task.Description] = true
					assert.NotEmpty(t, task.Id)
					assert.NotNil(t, task.CreatedAt)
					assert.NotNil(t, task.UpdatedAt)
				}
				
				// Verify all expected descriptions are present
				for _, expectedDesc := range tt.setupTasks {
					assert.True(t, taskDescriptions[expectedDesc], 
						"Expected task description '%s' not found", expectedDesc)
				}
			}
		})
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	tests := []struct {
		name            string
		setupTask       bool
		deleteId        string
		expectedSuccess bool
		expectedMessage string
	}{
		{
			name:            "delete_existing_task",
			setupTask:       true,
			deleteId:        "1",
			expectedSuccess: true,
			expectedMessage: "Task deleted successfully",
		},
		{
			name:            "delete_nonexistent_task",
			setupTask:       false,
			deleteId:        "999",
			expectedSuccess: false,
			expectedMessage: "task not found",
		},
		{
			name:            "delete_empty_id",
			setupTask:       true,
			deleteId:        "",
			expectedSuccess: false,
			expectedMessage: "task not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			taskStore := store.New()
			service := NewTaskService(taskStore)
			ctx := context.Background()
			
			if tt.setupTask {
				taskStore.CreateTask("Test task for deletion")
			}
			
			req := connect.NewRequest(&taskv1.DeleteTaskRequest{
				Id: tt.deleteId,
			})
			
			// Execute
			resp, err := service.DeleteTask(ctx, req)
			
			// Assert
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			
			response := resp.Msg
			assert.Equal(t, tt.expectedSuccess, response.Success)
			assert.Equal(t, tt.expectedMessage, response.Message)
			
			// If deletion was successful, verify task is actually gone
			if tt.expectedSuccess {
				_, err := taskStore.GetTask(tt.deleteId)
				assert.Error(t, err)
				assert.Equal(t, store.ErrTaskNotFound, err)
			}
		})
	}
}

func TestTaskService_Integration(t *testing.T) {
	// Full integration test covering the complete workflow
	taskStore := store.New()
	service := NewTaskService(taskStore)
	ctx := context.Background()
	
	// 1. Initially no tasks
	getAllReq := connect.NewRequest(&taskv1.GetAllTasksRequest{})
	getAllResp, err := service.GetAllTasks(ctx, getAllReq)
	require.NoError(t, err)
	assert.Empty(t, getAllResp.Msg.Tasks)
	
	// 2. Create first task
	createReq1 := connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "First task",
	})
	createResp1, err := service.CreateTask(ctx, createReq1)
	require.NoError(t, err)
	task1 := createResp1.Msg.Task
	assert.Equal(t, "1", task1.Id)
	assert.Equal(t, "First task", task1.Description)
	
	// 3. Create second task
	createReq2 := connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "Second task",
	})
	createResp2, err := service.CreateTask(ctx, createReq2)
	require.NoError(t, err)
	task2 := createResp2.Msg.Task
	assert.Equal(t, "2", task2.Id)
	assert.Equal(t, "Second task", task2.Description)
	
	// 4. Get all tasks - should have 2
	getAllResp, err = service.GetAllTasks(ctx, getAllReq)
	require.NoError(t, err)
	assert.Len(t, getAllResp.Msg.Tasks, 2)
	
	// 5. Delete first task
	deleteReq := connect.NewRequest(&taskv1.DeleteTaskRequest{
		Id: task1.Id,
	})
	deleteResp, err := service.DeleteTask(ctx, deleteReq)
	require.NoError(t, err)
	assert.True(t, deleteResp.Msg.Success)
	
	// 6. Get all tasks - should have 1
	getAllResp, err = service.GetAllTasks(ctx, getAllReq)
	require.NoError(t, err)
	require.Len(t, getAllResp.Msg.Tasks, 1)
	assert.Equal(t, task2.Id, getAllResp.Msg.Tasks[0].Id)
	
	// 7. Try to delete already deleted task
	deleteResp, err = service.DeleteTask(ctx, deleteReq)
	require.NoError(t, err)
	assert.False(t, deleteResp.Msg.Success)
	assert.Contains(t, deleteResp.Msg.Message, "not found")
}

func TestTaskService_ConcurrentOperations(t *testing.T) {
	taskStore := store.New()
	service := NewTaskService(taskStore)
	ctx := context.Background()
	
	const numGoroutines = 10
	const numOperationsPerGoroutine = 20
	
	done := make(chan bool, numGoroutines)
	
	// Run concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineId int) {
			for j := 0; j < numOperationsPerGoroutine; j++ {
				// Create task
				createReq := connect.NewRequest(&taskv1.CreateTaskRequest{
					Description: "Concurrent task",
				})
				_, err := service.CreateTask(ctx, createReq)
				assert.NoError(t, err)
				
				// Get all tasks
				getAllReq := connect.NewRequest(&taskv1.GetAllTasksRequest{})
				_, err = service.GetAllTasks(ctx, getAllReq)
				assert.NoError(t, err)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify final state
	getAllReq := connect.NewRequest(&taskv1.GetAllTasksRequest{})
	getAllResp, err := service.GetAllTasks(ctx, getAllReq)
	require.NoError(t, err)
	assert.Len(t, getAllResp.Msg.Tasks, numGoroutines*numOperationsPerGoroutine)
}

func BenchmarkTaskService_CreateTask(b *testing.B) {
	taskStore := store.New()
	service := NewTaskService(taskStore)
	ctx := context.Background()
	
	req := connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "Benchmark task",
	})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CreateTask(ctx, req)
	}
}

func BenchmarkTaskService_GetAllTasks(b *testing.B) {
	taskStore := store.New()
	service := NewTaskService(taskStore)
	ctx := context.Background()
	
	// Pre-populate with tasks
	for i := 0; i < 1000; i++ {
		taskStore.CreateTask("Benchmark task")
	}
	
	req := connect.NewRequest(&taskv1.GetAllTasksRequest{})
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetAllTasks(ctx, req)
	}
}