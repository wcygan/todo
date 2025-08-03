package store

import (
	"testing"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	store := New()
	
	assert.NotNil(t, store)
	assert.Equal(t, int64(1), store.nextID)
	assert.Empty(t, store.tasks)
}

func TestTaskStore_CreateTask(t *testing.T) {
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
			
			task := store.CreateTask(tt.description)
			
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
	store := New()
	
	// Create multiple tasks concurrently
	const numTasks = 10
	results := make(chan string, numTasks)
	
	for i := 0; i < numTasks; i++ {
		go func(i int) {
			task := store.CreateTask("Task " + string(rune(i+'0')))
			results <- task.Id
		}(i)
	}
	
	// Collect all IDs
	ids := make(map[string]bool)
	for i := 0; i < numTasks; i++ {
		id := <-results
		assert.False(t, ids[id], "Duplicate ID found: %s", id)
		ids[id] = true
	}
	
	assert.Len(t, ids, numTasks)
}

func TestTaskStore_GetTask(t *testing.T) {
	store := New()
	
	// Create a task
	originalTask := store.CreateTask("Test task")
	
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
			task, err := store.GetTask(tt.id)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrTaskNotFound, err)
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
	store := New()
	
	// Initially empty
	tasks := store.ListTasks()
	assert.Empty(t, tasks)
	
	// Add some tasks
	task1 := store.CreateTask("Task 1")
	task2 := store.CreateTask("Task 2")
	task3 := store.CreateTask("Task 3")
	
	tasks = store.ListTasks()
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
	store := New()
	
	// Create a task
	originalTask := store.CreateTask("Original description")
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
			task, err := store.UpdateTask(tt.id, tt.description, tt.completed)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrTaskNotFound, err)
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
	store := New()
	
	// Create a task
	task := store.CreateTask("Test task")
	
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
			err := store.DeleteTask(tt.id)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, ErrTaskNotFound, err)
			} else {
				assert.NoError(t, err)
				
				// Verify task is actually deleted
				_, getErr := store.GetTask(tt.id)
				assert.Error(t, getErr)
				assert.Equal(t, ErrTaskNotFound, getErr)
			}
		})
	}
}

func TestTaskStore_ThreadSafety(t *testing.T) {
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
				task := store.CreateTask("Task from goroutine")
				
				// Read tasks
				store.ListTasks()
				
				// Try to get the task
				store.GetTask(task.Id)
				
				// Update the task
				store.UpdateTask(task.Id, "Updated", true)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify final state
	tasks := store.ListTasks()
	assert.Len(t, tasks, numGoroutines*numOperations)
}

func BenchmarkTaskStore_CreateTask(b *testing.B) {
	store := New()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.CreateTask("Benchmark task")
	}
}

func BenchmarkTaskStore_ListTasks(b *testing.B) {
	store := New()
	
	// Pre-populate with tasks
	for i := 0; i < 1000; i++ {
		store.CreateTask("Task")
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.ListTasks()
	}
}