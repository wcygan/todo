package service

import (
	"context"
	"testing"

	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/wcygan/todo/backend/internal/errors"
)

// MockTaskRepository is a mock implementation of TaskRepository
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) CreateTask(ctx context.Context, description string) (*taskv1.Task, error) {
	args := m.Called(ctx, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*taskv1.Task), args.Error(1)
}

func (m *MockTaskRepository) GetTask(ctx context.Context, id string) (*taskv1.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*taskv1.Task), args.Error(1)
}

func (m *MockTaskRepository) ListTasks(ctx context.Context) ([]*taskv1.Task, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*taskv1.Task), args.Error(1)
}

func (m *MockTaskRepository) UpdateTask(ctx context.Context, id, description string, completed bool) (*taskv1.Task, error) {
	args := m.Called(ctx, id, description, completed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*taskv1.Task), args.Error(1)
}

func (m *MockTaskRepository) DeleteTask(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestNewTaskService(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	service := NewTaskService(mockRepo)
	
	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.repo)
}

func TestTaskService_CreateTask(t *testing.T) {
	tests := []struct {
		name        string
		description string
		mockSetup   func(*MockTaskRepository)
		wantErr     bool
		errCode     errors.ErrorCode
	}{
		{
			name:        "successful_creation",
			description: "Test task",
			mockSetup: func(m *MockTaskRepository) {
				task := &taskv1.Task{
					Id:          "1",
					Description: "Test task",
					Completed:   false,
					CreatedAt:   timestamppb.Now(),
					UpdatedAt:   timestamppb.Now(),
				}
				m.On("CreateTask", mock.Anything, "Test task").Return(task, nil)
			},
			wantErr: false,
		},
		{
			name:        "empty_description",
			description: "",
			mockSetup:   func(m *MockTaskRepository) {},
			wantErr:     true,
			errCode:     errors.CodeValidation,
		},
		{
			name:        "repository_error",
			description: "Test task",
			mockSetup: func(m *MockTaskRepository) {
				m.On("CreateTask", mock.Anything, "Test task").Return(nil, assert.AnError)
			},
			wantErr: true,
			errCode: errors.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			tt.mockSetup(mockRepo)
			
			service := NewTaskService(mockRepo)
			ctx := context.Background()
			
			task, err := service.CreateTask(ctx, tt.description)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, task)
				
				var appErr *errors.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, tt.description, task.Description)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_GetTask(t *testing.T) {
	tests := []struct {
		name      string
		taskID    string
		mockSetup func(*MockTaskRepository)
		wantErr   bool
		errCode   errors.ErrorCode
	}{
		{
			name:   "successful_get",
			taskID: "1",
			mockSetup: func(m *MockTaskRepository) {
				task := &taskv1.Task{
					Id:          "1",
					Description: "Test task",
					Completed:   false,
				}
				m.On("GetTask", mock.Anything, "1").Return(task, nil)
			},
			wantErr: false,
		},
		{
			name:      "empty_id",
			taskID:    "",
			mockSetup: func(m *MockTaskRepository) {},
			wantErr:   true,
			errCode:   errors.CodeValidation,
		},
		{
			name:   "task_not_found",
			taskID: "999",
			mockSetup: func(m *MockTaskRepository) {
				m.On("GetTask", mock.Anything, "999").Return(nil, errors.NotFound("task", "999"))
			},
			wantErr: true,
			errCode: errors.CodeNotFound,
		},
		{
			name:   "repository_error",
			taskID: "1",
			mockSetup: func(m *MockTaskRepository) {
				m.On("GetTask", mock.Anything, "1").Return(nil, assert.AnError)
			},
			wantErr: true,
			errCode: errors.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			tt.mockSetup(mockRepo)
			
			service := NewTaskService(mockRepo)
			ctx := context.Background()
			
			task, err := service.GetTask(ctx, tt.taskID)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, task)
				
				var appErr *errors.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				require.NotNil(t, task)
				assert.Equal(t, tt.taskID, task.Id)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_ListTasks(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(*MockTaskRepository)
		wantErr   bool
		errCode   errors.ErrorCode
	}{
		{
			name: "successful_list",
			mockSetup: func(m *MockTaskRepository) {
				tasks := []*taskv1.Task{
					{Id: "1", Description: "Task 1"},
					{Id: "2", Description: "Task 2"},
				}
				m.On("ListTasks", mock.Anything).Return(tasks, nil)
			},
			wantErr: false,
		},
		{
			name: "empty_list",
			mockSetup: func(m *MockTaskRepository) {
				m.On("ListTasks", mock.Anything).Return([]*taskv1.Task{}, nil)
			},
			wantErr: false,
		},
		{
			name: "repository_error",
			mockSetup: func(m *MockTaskRepository) {
				m.On("ListTasks", mock.Anything).Return(nil, assert.AnError)
			},
			wantErr: true,
			errCode: errors.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			tt.mockSetup(mockRepo)
			
			service := NewTaskService(mockRepo)
			ctx := context.Background()
			
			tasks, err := service.ListTasks(ctx)
			
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, tasks)
				
				var appErr *errors.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tasks)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	tests := []struct {
		name      string
		taskID    string
		mockSetup func(*MockTaskRepository)
		wantErr   bool
		errCode   errors.ErrorCode
	}{
		{
			name:   "successful_delete",
			taskID: "1",
			mockSetup: func(m *MockTaskRepository) {
				m.On("DeleteTask", mock.Anything, "1").Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "empty_id",
			taskID:    "",
			mockSetup: func(m *MockTaskRepository) {},
			wantErr:   true,
			errCode:   errors.CodeValidation,
		},
		{
			name:   "task_not_found",
			taskID: "999",
			mockSetup: func(m *MockTaskRepository) {
				m.On("DeleteTask", mock.Anything, "999").Return(errors.NotFound("task", "999"))
			},
			wantErr: true,
			errCode: errors.CodeNotFound,
		},
		{
			name:   "repository_error",
			taskID: "1",
			mockSetup: func(m *MockTaskRepository) {
				m.On("DeleteTask", mock.Anything, "1").Return(assert.AnError)
			},
			wantErr: true,
			errCode: errors.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockTaskRepository{}
			tt.mockSetup(mockRepo)
			
			service := NewTaskService(mockRepo)
			ctx := context.Background()
			
			err := service.DeleteTask(ctx, tt.taskID)
			
			if tt.wantErr {
				require.Error(t, err)
				
				var appErr *errors.Error
				require.True(t, errors.As(err, &appErr))
				assert.Equal(t, tt.errCode, appErr.Code)
			} else {
				require.NoError(t, err)
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}