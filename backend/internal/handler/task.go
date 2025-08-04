package handler

import (
	"context"

	"connectrpc.com/connect"
	taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
	taskconnect "buf.build/gen/go/wcygan/todo/connectrpc/go/task/v1/taskv1connect"

	"github.com/wcygan/todo/backend/internal/errors"
	"github.com/wcygan/todo/backend/internal/service"
)

// TaskHandler implements the TaskService ConnectRPC interface
type TaskHandler struct {
	service *service.TaskService
}

// NewTaskHandler creates a new TaskHandler instance
func NewTaskHandler(service *service.TaskService) *TaskHandler {
	return &TaskHandler{
		service: service,
	}
}

// CreateTask handles task creation requests
func (h *TaskHandler) CreateTask(
	ctx context.Context,
	req *connect.Request[taskv1.CreateTaskRequest],
) (*connect.Response[taskv1.CreateTaskResponse], error) {
	task, err := h.service.CreateTask(ctx, req.Msg.Description)
	if err != nil {
		return nil, errors.ToConnectError(err)
	}

	return connect.NewResponse(&taskv1.CreateTaskResponse{
		Task: task,
	}), nil
}

// GetTask handles requests to retrieve a single task by ID
func (h *TaskHandler) GetTask(
	ctx context.Context,
	req *connect.Request[taskv1.GetTaskRequest],
) (*connect.Response[taskv1.GetTaskResponse], error) {
	task, err := h.service.GetTask(ctx, req.Msg.Id)
	if err != nil {
		return nil, errors.ToConnectError(err)
	}

	return connect.NewResponse(&taskv1.GetTaskResponse{
		Task: task,
	}), nil
}

// GetAllTasks handles requests to retrieve all tasks
func (h *TaskHandler) GetAllTasks(
	ctx context.Context,
	req *connect.Request[taskv1.GetAllTasksRequest],
) (*connect.Response[taskv1.GetAllTasksResponse], error) {
	tasks, err := h.service.ListTasks(ctx)
	if err != nil {
		return nil, errors.ToConnectError(err)
	}

	return connect.NewResponse(&taskv1.GetAllTasksResponse{
		Tasks: tasks,
	}), nil
}

// UpdateTask handles task update requests
func (h *TaskHandler) UpdateTask(
	ctx context.Context,
	req *connect.Request[taskv1.UpdateTaskRequest],
) (*connect.Response[taskv1.UpdateTaskResponse], error) {
	task, err := h.service.UpdateTask(ctx, req.Msg.Id, req.Msg.Description, req.Msg.Completed)
	if err != nil {
		return nil, errors.ToConnectError(err)
	}

	return connect.NewResponse(&taskv1.UpdateTaskResponse{
		Task: task,
	}), nil
}

// DeleteTask handles task deletion requests
func (h *TaskHandler) DeleteTask(
	ctx context.Context,
	req *connect.Request[taskv1.DeleteTaskRequest],
) (*connect.Response[taskv1.DeleteTaskResponse], error) {
	err := h.service.DeleteTask(ctx, req.Msg.Id)
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

// Verify that TaskHandler implements the interface
var _ taskconnect.TaskServiceHandler = (*TaskHandler)(nil)