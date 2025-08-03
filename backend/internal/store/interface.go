package store

import (
	"context"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
)

// TaskRepository defines the interface for task storage operations
type TaskRepository interface {
	// CreateTask creates a new task with the given description
	CreateTask(ctx context.Context, description string) (*taskv1.Task, error)
	
	// GetTask retrieves a task by ID
	GetTask(ctx context.Context, id string) (*taskv1.Task, error)
	
	// ListTasks returns all tasks in the store
	ListTasks(ctx context.Context) ([]*taskv1.Task, error)
	
	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, id, description string, completed bool) (*taskv1.Task, error)
	
	// DeleteTask removes a task by ID
	DeleteTask(ctx context.Context, id string) error
}