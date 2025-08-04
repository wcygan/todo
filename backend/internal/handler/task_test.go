package handler

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wcygan/todo/backend/internal/service"
	"github.com/wcygan/todo/backend/test/testutil"
)

func TestNewTaskHandler(t *testing.T) {
	taskStore := testutil.NewMockStore()
	taskService := service.NewTaskService(taskStore)
	handler := NewTaskHandler(taskService)
	
	assert.NotNil(t, handler)
	assert.Equal(t, taskService, handler.service)
}

func TestTaskHandler_CreateTask(t *testing.T) {
	tests := []struct {
		name        string
		description string
		wantErr     bool
		wantErrCode connect.Code
	}{
		{
			name:        "create_valid_task",
			description: "Test task description",
			wantErr:     false,
		},
		{
			name:        "create_empty_description",
			description: "",
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
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
			taskStore := testutil.NewMockStore()
			taskService := service.NewTaskService(taskStore)
			handler := NewTaskHandler(taskService)
			ctx := context.Background()
			
			req := connect.NewRequest(&taskv1.CreateTaskRequest{
				Description: tt.description,
			})
			
			// Execute
			resp, err := handler.CreateTask(ctx, req)
			
			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				if tt.wantErrCode != 0 {
					var connectErr *connect.Error
					require.ErrorAs(t, err, &connectErr)
					assert.Equal(t, tt.wantErrCode, connectErr.Code())
				}
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

func TestTaskHandler_GetAllTasks(t *testing.T) {
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
			taskStore := testutil.NewMockStore()
			taskService := service.NewTaskService(taskStore)
			handler := NewTaskHandler(taskService)
			ctx := context.Background()
			
			// Create setup tasks
			for _, desc := range tt.setupTasks {
				_, err := taskStore.CreateTask(ctx, desc)
				require.NoError(t, err)
			}
			
			req := connect.NewRequest(&taskv1.GetAllTasksRequest{})
			
			// Execute
			resp, err := handler.GetAllTasks(ctx, req)
			
			// Assert
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			assert.Len(t, resp.Msg.Tasks, tt.expectedCount)
			
			// Verify task content
			if tt.expectedCount > 0 {
				for i, task := range resp.Msg.Tasks {
					assert.NotEmpty(t, task.Id)
					assert.Equal(t, tt.setupTasks[i], task.Description)
					assert.False(t, task.Completed)
					assert.NotNil(t, task.CreatedAt)
					assert.NotNil(t, task.UpdatedAt)
				}
			}
		})
	}
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	tests := []struct {
		name        string
		setupTask   bool
		taskID      string
		expectError bool
	}{
		{
			name:        "delete_existing_task",
			setupTask:   true,
			taskID:      "1",
			expectError: false,
		},
		{
			name:        "delete_nonexistent_task",
			setupTask:   false,
			taskID:      "999",
			expectError: false, // Handler returns success=false, not an error
		},
		{
			name:        "delete_empty_id",
			setupTask:   false,
			taskID:      "",
			expectError: false, // Service validation will handle this
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			taskStore := testutil.NewMockStore()
			taskService := service.NewTaskService(taskStore)
			handler := NewTaskHandler(taskService)
			ctx := context.Background()
			
			if tt.setupTask {
				_, err := taskStore.CreateTask(ctx, "Test task")
				require.NoError(t, err)
			}
			
			req := connect.NewRequest(&taskv1.DeleteTaskRequest{
				Id: tt.taskID,
			})
			
			// Execute
			resp, err := handler.DeleteTask(ctx, req)
			
			// Assert
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Msg)
			
			if tt.setupTask && tt.taskID == "1" {
				assert.True(t, resp.Msg.Success)
				assert.Contains(t, resp.Msg.Message, "successfully")
			} else {
				assert.False(t, resp.Msg.Success)
				assert.NotEmpty(t, resp.Msg.Message)
			}
		})
	}
}

func TestTaskHandler_CreateTask_ValidationErrors(t *testing.T) {
	taskStore := testutil.NewMockStore()
	taskService := service.NewTaskService(taskStore)
	handler := NewTaskHandler(taskService)
	ctx := context.Background()
	
	// Test empty description validation
	req := connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "",
	})
	
	resp, err := handler.CreateTask(ctx, req)
	
	assert.Error(t, err)
	assert.Nil(t, resp)
	
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

func TestTaskHandler_DeleteTask_WithStoreError(t *testing.T) {
	taskStore := testutil.NewMockStore()
	taskService := service.NewTaskService(taskStore)
	handler := NewTaskHandler(taskService)
	ctx := context.Background()
	
	// Try to delete a non-existent task
	req := connect.NewRequest(&taskv1.DeleteTaskRequest{
		Id: "nonexistent",
	})
	
	resp, err := handler.DeleteTask(ctx, req)
	
	// The handler should not return an error, but success should be false
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Msg.Success)
	assert.Contains(t, resp.Msg.Message, "not found")
}

func TestTaskHandler_IntegrationTest(t *testing.T) {
	// Setup
	taskStore := testutil.NewMockStore()
	taskService := service.NewTaskService(taskStore)
	handler := NewTaskHandler(taskService)
	ctx := context.Background()
	
	// Create a task
	createReq := connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "Integration test task",
	})
	
	createResp, err := handler.CreateTask(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	
	taskID := createResp.Msg.Task.Id
	
	// Get all tasks - should have one
	getAllReq := connect.NewRequest(&taskv1.GetAllTasksRequest{})
	getAllResp, err := handler.GetAllTasks(ctx, getAllReq)
	require.NoError(t, err)
	require.Len(t, getAllResp.Msg.Tasks, 1)
	assert.Equal(t, taskID, getAllResp.Msg.Tasks[0].Id)
	
	// Delete the task
	deleteReq := connect.NewRequest(&taskv1.DeleteTaskRequest{
		Id: taskID,
	})
	deleteResp, err := handler.DeleteTask(ctx, deleteReq)
	require.NoError(t, err)
	assert.True(t, deleteResp.Msg.Success)
	
	// Get all tasks - should be empty
	getAllResp2, err := handler.GetAllTasks(ctx, getAllReq)
	require.NoError(t, err)
	assert.Empty(t, getAllResp2.Msg.Tasks)
}

func TestTaskHandler_ContextCancellation(t *testing.T) {
	taskStore := testutil.NewMockStore()
	taskService := service.NewTaskService(taskStore)
	handler := NewTaskHandler(taskService)
	
	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	
	req := connect.NewRequest(&taskv1.CreateTaskRequest{
		Description: "Test task",
	})
	
	_, err := handler.CreateTask(ctx, req)
	assert.Error(t, err)
	
	var connectErr *connect.Error
	require.ErrorAs(t, err, &connectErr)
	assert.Equal(t, connect.CodeInternal, connectErr.Code())
}