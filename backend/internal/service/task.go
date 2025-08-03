package service

import (
	"context"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"

	"github.com/wcygan/todo/backend/internal/errors"
	"github.com/wcygan/todo/backend/internal/store"
)

// TaskService handles business logic for task operations
type TaskService struct {
	repo store.TaskRepository
}

// NewTaskService creates a new TaskService instance
func NewTaskService(repo store.TaskRepository) *TaskService {
	return &TaskService{
		repo: repo,
	}
}

// CreateTask creates a new task with validation
func (s *TaskService) CreateTask(ctx context.Context, description string) (*taskv1.Task, error) {
	// Validate input
	if description == "" {
		return nil, errors.Validation("description", "description cannot be empty")
	}

	// Create task
	task, err := s.repo.CreateTask(ctx, description)
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to create task")
	}

	return task, nil
}

// GetTask retrieves a task by ID
func (s *TaskService) GetTask(ctx context.Context, id string) (*taskv1.Task, error) {
	if id == "" {
		return nil, errors.Validation("id", "task ID cannot be empty")
	}

	task, err := s.repo.GetTask(ctx, id)
	if err != nil {
		// Pass through not found errors, wrap others
		if errors.IsNotFound(err) {
			return nil, err
		}
		return nil, errors.InternalWrap(err, "failed to get task")
	}

	return task, nil
}

// ListTasks returns all tasks
func (s *TaskService) ListTasks(ctx context.Context) ([]*taskv1.Task, error) {
	tasks, err := s.repo.ListTasks(ctx)
	if err != nil {
		return nil, errors.InternalWrap(err, "failed to list tasks")
	}

	return tasks, nil
}

// UpdateTask updates an existing task
func (s *TaskService) UpdateTask(ctx context.Context, id, description string, completed bool) (*taskv1.Task, error) {
	if id == "" {
		return nil, errors.Validation("id", "task ID cannot be empty")
	}

	task, err := s.repo.UpdateTask(ctx, id, description, completed)
	if err != nil {
		// Pass through not found errors, wrap others
		if errors.IsNotFound(err) {
			return nil, err
		}
		return nil, errors.InternalWrap(err, "failed to update task")
	}

	return task, nil
}

// DeleteTask removes a task by ID
func (s *TaskService) DeleteTask(ctx context.Context, id string) error {
	if id == "" {
		return errors.Validation("id", "task ID cannot be empty")
	}

	err := s.repo.DeleteTask(ctx, id)
	if err != nil {
		// Pass through not found errors, wrap others
		if errors.IsNotFound(err) {
			return err
		}
		return errors.InternalWrap(err, "failed to delete task")
	}

	return nil
}