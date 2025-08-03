package store

import (
	"errors"
	"strconv"
	"sync"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrTaskNotFound = errors.New("task not found")
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
func (s *TaskStore) CreateTask(description string) *taskv1.Task {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	return task
}

// GetTask retrieves a task by ID
func (s *TaskStore) GetTask(id string) (*taskv1.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// ListTasks returns all tasks in the store
func (s *TaskStore) ListTasks() []*taskv1.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*taskv1.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// UpdateTask updates an existing task
func (s *TaskStore) UpdateTask(id, description string, completed bool) (*taskv1.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, ErrTaskNotFound
	}

	if description != "" {
		task.Description = description
	}
	task.Completed = completed
	task.UpdatedAt = timestamppb.Now()

	return task, nil
}

// DeleteTask removes a task by ID
func (s *TaskStore) DeleteTask(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[id]; !exists {
		return ErrTaskNotFound
	}

	delete(s.tasks, id)
	return nil
}