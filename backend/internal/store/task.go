package store

import (
	"context"
	"strconv"
	"sync"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/wcygan/todo/backend/internal/errors"
)

// TaskStore provides thread-safe in-memory storage for tasks
type TaskStore struct {
	mu     sync.RWMutex
	tasks  map[string]*taskv1.Task
	nextID int64
}

// New creates a new TaskStore instance
func New() *TaskStore {
	return &TaskStore{
		tasks:  make(map[string]*taskv1.Task),
		nextID: 1,
	}
}

// CreateTask creates a new task with the given description
func (s *TaskStore) CreateTask(ctx context.Context, description string) (*taskv1.Task, error) {
	// Check for context cancellation before acquiring lock
	select {
	case <-ctx.Done():
		return nil, errors.InternalWrap(ctx.Err(), "context cancelled during task creation")
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check again after acquiring lock
	select {
	case <-ctx.Done():
		return nil, errors.InternalWrap(ctx.Err(), "context cancelled during task creation")
	default:
	}

	now := timestamppb.Now()
	task := &taskv1.Task{
		Id:          strconv.FormatInt(s.nextID, 10),
		Description: description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.tasks[strconv.FormatInt(s.nextID, 10)] = task
	s.nextID++

	return task, nil
}

// GetTask retrieves a task by ID
func (s *TaskStore) GetTask(ctx context.Context, id string) (*taskv1.Task, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, errors.InternalWrap(ctx.Err(), "context cancelled during task retrieval")
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, errors.NotFound("task", id)
	}

	return task, nil
}

// ListTasks returns all tasks in the store
func (s *TaskStore) ListTasks(ctx context.Context) ([]*taskv1.Task, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, errors.InternalWrap(ctx.Err(), "context cancelled during task listing")
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*taskv1.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		// Check cancellation during iteration for large datasets
		select {
		case <-ctx.Done():
			return nil, errors.InternalWrap(ctx.Err(), "context cancelled during task listing")
		default:
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask updates an existing task
func (s *TaskStore) UpdateTask(ctx context.Context, id, description string, completed bool) (*taskv1.Task, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, errors.InternalWrap(ctx.Err(), "context cancelled during task update")
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, errors.NotFound("task", id)
	}

	if description != "" {
		task.Description = description
	}
	task.Completed = completed
	task.UpdatedAt = timestamppb.Now()

	return task, nil
}

// DeleteTask removes a task by ID
func (s *TaskStore) DeleteTask(ctx context.Context, id string) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return errors.InternalWrap(ctx.Err(), "context cancelled during task deletion")
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return errors.NotFound("task", id)
	}

	delete(s.tasks, id)
	return nil
}

// Verify that TaskStore implements the TaskRepository interface
var _ TaskRepository = (*TaskStore)(nil)