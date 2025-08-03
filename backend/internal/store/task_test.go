package store

import (
	"context"
	"testing"
	"time"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	
	"github.com/wcygan/todo/backend/internal/errors"
)

func TestNew(t *testing.T) {
	store := New()
	
	assert.NotNil(t, store)
	assert.Equal(t, int64(1), store.nextID)
	assert.Empty(t, store.tasks)
}

func TestTaskStore_CreateTask(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "create_simple_task",
			description: "Test task",
			want:        "Test task",
		},
		{
			name:        "create_empty_description",
			description: "",
			want:        "",
		},
		{
			name:        "create_long_description",
			description: "This is a very long task description that should be handled properly",
			want:        "This is a very long task description that should be handled properly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := New()
			
			task, err := store.CreateTask(ctx, tt.description)
			
			require.NoError(t, err)
			require.NotNil(t, task)
			assert.Equal(t, "1", task.Id)
			assert.Equal(t, tt.want, task.Description)
			assert.False(t, task.Completed)
			assert.NotNil(t, task.CreatedAt)
			assert.NotNil(t, task.UpdatedAt)
			assert.Equal(t, task.CreatedAt, task.UpdatedAt)
		})
	}
}

func TestTaskStore_CreateTask_Concurrent(t *testing.T) {
	ctx := context.Background()
	store := New()
	
	// Create multiple tasks concurrently
	const numTasks = 10
	results := make(chan string, numTasks)
	errors := make(chan error, numTasks)
	
	for i := 0; i < numTasks; i++ {
		go func(i int) {
			task, err := store.CreateTask(ctx, "Task "+string(rune(i+'0')))
			if err != nil {
				errors <- err
				return
			}
			results <- task.Id
		}(i)
	}
	
	// Collect all IDs
	ids := make(map[string]bool)
	for i := 0; i < numTasks; i++ {
		select {
		case id := <-results:
			assert.False(t, ids[id], "Duplicate ID found: %s", id)
			ids[id] = true
		case err := <-errors:
			t.Fatalf("Unexpected error: %v", err)
		}
	}
	
	assert.Len(t, ids, numTasks)
}

func TestTaskStore_GetTask(t *testing.T) {
	ctx := context.Background()
	store := New()
	
	// Create a task
	originalTask, err := store.CreateTask(ctx, "Test task")
	require.NoError(t, err)
	
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "get_existing_task",
			id:      originalTask.Id,
			wantErr: false,
		},
		{
			name:    "get_nonexistent_task",
			id:      "999",
			wantErr: true,
		},
		{
			name:    "get_empty_id",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := store.GetTask(ctx, tt.id)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFound(err))
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, originalTask.Id, task.Id)
				assert.Equal(t, originalTask.Description, task.Description)
			}
		})
	}
}

func TestTaskStore_ListTasks(t *testing.T) {
	ctx := context.Background()
	store := New()
	
	// Initially empty
	tasks, err := store.ListTasks(ctx)
	require.NoError(t, err)
	assert.Empty(t, tasks)
	
	// Add some tasks
	task1, err := store.CreateTask(ctx, "Task 1")
	require.NoError(t, err)
	task2, err := store.CreateTask(ctx, "Task 2")
	require.NoError(t, err)
	task3, err := store.CreateTask(ctx, "Task 3")
	require.NoError(t, err)
	
	tasks, err = store.ListTasks(ctx)
	require.NoError(t, err)
	require.Len(t, tasks, 3)
	
	// Check that all tasks are present (order may vary due to map iteration)
	taskMap := make(map[string]*taskv1.Task)
	for _, task := range tasks {
		taskMap[task.Id] = task
	}
	
	assert.Equal(t, task1.Description, taskMap[task1.Id].Description)
	assert.Equal(t, task2.Description, taskMap[task2.Id].Description)
	assert.Equal(t, task3.Description, taskMap[task3.Id].Description)
}

func TestTaskStore_UpdateTask(t *testing.T) {
	ctx := context.Background()
	store := New()
	
	// Create a task
	originalTask, err := store.CreateTask(ctx, "Original description")
	require.NoError(t, err)
	originalUpdatedAt := originalTask.UpdatedAt
	
	tests := []struct {
		name        string
		id          string
		description string
		completed   bool
		wantErr     bool
	}{
		{
			name:        "update_description",
			id:          originalTask.Id,
			description: "Updated description",
			completed:   false,
			wantErr:     false,
		},
		{
			name:        "update_completed_status",
			id:          originalTask.Id,
			description: "",
			completed:   true,
			wantErr:     false,
		},
		{
			name:        "update_nonexistent_task",
			id:          "999",
			description: "New description",
			completed:   false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := store.UpdateTask(ctx, tt.id, tt.description, tt.completed)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFound(err))
				assert.Nil(t, task)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, tt.id, task.Id)
				assert.Equal(t, tt.completed, task.Completed)
				
				if tt.description != "" {
					assert.Equal(t, tt.description, task.Description)
				}
				
				// UpdatedAt should be newer
				assert.True(t, task.UpdatedAt.AsTime().After(originalUpdatedAt.AsTime()) || 
					task.UpdatedAt.AsTime().Equal(originalUpdatedAt.AsTime()))
			}
		})
	}
}

func TestTaskStore_DeleteTask(t *testing.T) {
	ctx := context.Background()
	store := New()
	
	// Create a task
	task, err := store.CreateTask(ctx, "Test task")
	require.NoError(t, err)
	
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "delete_existing_task",
			id:      task.Id,
			wantErr: false,
		},
		{
			name:    "delete_nonexistent_task",
			id:      "999",
			wantErr: true,
		},
		{
			name:    "delete_already_deleted_task",
			id:      task.Id,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.DeleteTask(ctx, tt.id)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.IsNotFound(err))
			} else {
				assert.NoError(t, err)
				
				// Verify task is actually deleted
				_, getErr := store.GetTask(ctx, tt.id)
				assert.Error(t, getErr)
				assert.True(t, errors.IsNotFound(getErr))
			}
		})
	}
}

func TestTaskStore_ContextCancellation(t *testing.T) {
	store := New()
	
	// Test context cancellation for each method
	t.Run("CreateTask_cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := store.CreateTask(ctx, "Test task")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cancelled")
	})
	
	t.Run("GetTask_cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := store.GetTask(ctx, "1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cancelled")
	})
	
	t.Run("ListTasks_cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := store.ListTasks(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cancelled")
	})
	
	t.Run("UpdateTask_cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := store.UpdateTask(ctx, "1", "Updated", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cancelled")
	})
	
	t.Run("DeleteTask_cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		err := store.DeleteTask(ctx, "1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cancelled")
	})
}

func TestTaskStore_ContextTimeout(t *testing.T) {
	store := New()
	
	// Test with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	
	time.Sleep(1 * time.Millisecond) // Ensure timeout occurs
	
	_, err := store.CreateTask(ctx, "Test task")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}

func TestTaskStore_ThreadSafety(t *testing.T) {
	ctx := context.Background()
	store := New()
	
	// Test concurrent operations
	const numGoroutines = 10
	const numOperations = 100
	
	done := make(chan bool, numGoroutines)
	
	// Create tasks concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				// Create task
				task, err := store.CreateTask(ctx, "Task from goroutine")
				if err != nil {
					continue
				}
				
				// Read tasks
				store.ListTasks(ctx)
				
				// Try to get the task
				store.GetTask(ctx, task.Id)
				
				// Update the task
				store.UpdateTask(ctx, task.Id, "Updated", true)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify final state
	tasks, err := store.ListTasks(ctx)
	require.NoError(t, err)
	assert.Len(t, tasks, numGoroutines*numOperations)
}

func BenchmarkTaskStore_CreateTask(b *testing.B) {
	ctx := context.Background()
	store := New()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.CreateTask(ctx, "Benchmark task")
	}
}

func BenchmarkTaskStore_ListTasks(b *testing.B) {
	ctx := context.Background()
	store := New()
	
	// Pre-populate with tasks
	for i := 0; i < 1000; i++ {
		store.CreateTask(ctx, "Task")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.ListTasks(ctx)
	}
}