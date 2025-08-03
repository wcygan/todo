package testutil

import (
	"testing"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/wcygan/todo/backend/internal/store"
)

// CreateTestTask creates a task for testing purposes
func CreateTestTask(description string) *taskv1.Task {
	now := timestamppb.Now()
	return &taskv1.Task{
		Id:          "test-id",
		Description: description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CreateTestTaskWithID creates a task with a specific ID for testing
func CreateTestTaskWithID(id, description string) *taskv1.Task {
	now := timestamppb.Now()
	return &taskv1.Task{
		Id:          id,
		Description: description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// CreateCompletedTestTask creates a completed task for testing
func CreateCompletedTestTask(id, description string) *taskv1.Task {
	now := timestamppb.Now()
	return &taskv1.Task{
		Id:          id,
		Description: description,
		Completed:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AssertTaskEquals compares two tasks for equality in tests
func AssertTaskEquals(t *testing.T, expected, actual *taskv1.Task) {
	t.Helper()
	
	require.NotNil(t, expected, "Expected task should not be nil")
	require.NotNil(t, actual, "Actual task should not be nil")
	
	assert.Equal(t, expected.Id, actual.Id, "Task IDs should match")
	assert.Equal(t, expected.Description, actual.Description, "Task descriptions should match")
	assert.Equal(t, expected.Completed, actual.Completed, "Task completion status should match")
	
	// For timestamps, we check they exist but don't compare exact values
	// as they may differ slightly due to timing
	assert.NotNil(t, actual.CreatedAt, "CreatedAt should not be nil")
	assert.NotNil(t, actual.UpdatedAt, "UpdatedAt should not be nil")
}

// AssertTaskListContains checks if a list contains a task with the given ID
func AssertTaskListContains(t *testing.T, tasks []*taskv1.Task, expectedID string) {
	t.Helper()
	
	found := false
	for _, task := range tasks {
		if task.Id == expectedID {
			found = true
			break
		}
	}
	assert.True(t, found, "Task list should contain task with ID: %s", expectedID)
}

// AssertTaskListDoesNotContain checks if a list does not contain a task with the given ID
func AssertTaskListDoesNotContain(t *testing.T, tasks []*taskv1.Task, expectedID string) {
	t.Helper()
	
	found := false
	for _, task := range tasks {
		if task.Id == expectedID {
			found = true
			break
		}
	}
	assert.False(t, found, "Task list should not contain task with ID: %s", expectedID)
}

// SetupTestStore creates a new store with predefined test data
func SetupTestStore(descriptions ...string) *store.TaskStore {
	testStore := store.New()
	
	for _, desc := range descriptions {
		testStore.CreateTask(desc)
	}
	
	return testStore
}

// MockStore is a simple mock implementation of the store interface for testing
type MockStore struct {
	tasks   map[string]*taskv1.Task
	nextID  int
	failing bool // Set to true to simulate errors
}

// NewMockStore creates a new mock store
func NewMockStore() *MockStore {
	return &MockStore{
		tasks:  make(map[string]*taskv1.Task),
		nextID: 1,
	}
}

// SetFailing makes the mock store return errors for all operations
func (m *MockStore) SetFailing(failing bool) {
	m.failing = failing
}

// CreateTask mock implementation
func (m *MockStore) CreateTask(description string) *taskv1.Task {
	if m.failing {
		return nil
	}
	
	task := CreateTestTaskWithID(string(rune(m.nextID+'0')), description)
	m.tasks[task.Id] = task
	m.nextID++
	return task
}

// GetTask mock implementation
func (m *MockStore) GetTask(id string) (*taskv1.Task, error) {
	if m.failing {
		return nil, store.ErrTaskNotFound
	}
	
	task, exists := m.tasks[id]
	if !exists {
		return nil, store.ErrTaskNotFound
	}
	return task, nil
}

// ListTasks mock implementation
func (m *MockStore) ListTasks() []*taskv1.Task {
	if m.failing {
		return nil
	}
	
	tasks := make([]*taskv1.Task, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// UpdateTask mock implementation
func (m *MockStore) UpdateTask(id, description string, completed bool) (*taskv1.Task, error) {
	if m.failing {
		return nil, store.ErrTaskNotFound
	}
	
	task, exists := m.tasks[id]
	if !exists {
		return nil, store.ErrTaskNotFound
	}
	
	if description != "" {
		task.Description = description
	}
	task.Completed = completed
	task.UpdatedAt = timestamppb.Now()
	
	return task, nil
}

// DeleteTask mock implementation
func (m *MockStore) DeleteTask(id string) error {
	if m.failing {
		return store.ErrTaskNotFound
	}
	
	if _, exists := m.tasks[id]; !exists {
		return store.ErrTaskNotFound
	}
	
	delete(m.tasks, id)
	return nil
}

// AddTask directly adds a task to the mock store (for test setup)
func (m *MockStore) AddTask(task *taskv1.Task) {
	m.tasks[task.Id] = task
}

// TaskCount returns the number of tasks in the mock store
func (m *MockStore) TaskCount() int {
	return len(m.tasks)
}

// Clear removes all tasks from the mock store
func (m *MockStore) Clear() {
	m.tasks = make(map[string]*taskv1.Task)
	m.nextID = 1
}