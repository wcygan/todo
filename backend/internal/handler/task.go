package handler

import (
	"context"

	"connectrpc.com/connect"
	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	taskconnect "buf.build/gen/go/wcygan/todo/connectrpc/go/task/v1/taskv1connect"

	"github.com/wcygan/todo/backend/internal/store"
)

// TaskService implements the TaskService ConnectRPC interface
type TaskService struct {
	store *store.TaskStore
}

// NewTaskService creates a new TaskService instance
func NewTaskService(store *store.TaskStore) *TaskService {
	return &TaskService{
		store: store,
	}
}

// CreateTask handles task creation requests
func (s *TaskService) CreateTask(
	ctx context.Context,
	req *connect.Request[taskv1.CreateTaskRequest],
) (*connect.Response[taskv1.CreateTaskResponse], error) {
	task := s.store.CreateTask(req.Msg.Description)

	return connect.NewResponse(&taskv1.CreateTaskResponse{
		Task: task,
	}), nil
}

// GetAllTasks handles requests to retrieve all tasks
func (s *TaskService) GetAllTasks(
	ctx context.Context,
	req *connect.Request[taskv1.GetAllTasksRequest],
) (*connect.Response[taskv1.GetAllTasksResponse], error) {
	tasks := s.store.ListTasks()

	return connect.NewResponse(&taskv1.GetAllTasksResponse{
		Tasks: tasks,
	}), nil
}

// DeleteTask handles task deletion requests
func (s *TaskService) DeleteTask(
	ctx context.Context,
	req *connect.Request[taskv1.DeleteTaskRequest],
) (*connect.Response[taskv1.DeleteTaskResponse], error) {
	err := s.store.DeleteTask(req.Msg.Id)
	if err != nil {
		return connect.NewResponse(&taskv1.DeleteTaskResponse{
			Success: false,
			Message: err.Error(),
		}), nil
	}

	return connect.NewResponse(&taskv1.DeleteTaskResponse{
		Success: true,
		Message: "Task deleted successfully",
	}), nil
}

// Verify that TaskService implements the interface
var _ taskconnect.TaskServiceHandler = (*TaskService)(nil)