# gRPC & Protocol Buffers Guide

This document explains how to work with our Protocol Buffers schema, Buf Schema Registry, and gRPC tooling for the Todo application.

## Overview

Our Todo application uses ConnectRPC, which provides both gRPC and HTTP/JSON APIs from a single Protocol Buffers definition. This gives us:

- **Type-safe APIs** with generated Go and TypeScript clients
- **HTTP/JSON REST API** for web compatibility  
- **Native gRPC** for high-performance service-to-service communication
- **Automatic code generation** from proto definitions

## Protocol Buffers Schema

### Schema Location
- **Proto files**: `proto/task/v1/task.proto`
- **Buf configuration**: `proto/buf.yaml`
- **Code generation**: `buf.gen.yaml`

### Current API Definition

```protobuf
service TaskService {
  rpc CreateTask(CreateTaskRequest) returns (CreateTaskResponse);
  rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
  rpc GetAllTasks(GetAllTasksRequest) returns (GetAllTasksResponse);
  rpc UpdateTask(UpdateTaskRequest) returns (UpdateTaskResponse);
  rpc DeleteTask(DeleteTaskRequest) returns (DeleteTaskResponse);
}
```

## Buf Schema Registry

### Generated Packages
Our proto definitions are published to Buf Schema Registry and generate:

- **Go**: `buf.build/gen/go/wcygan/todo/protocolbuffers/go`
- **Go ConnectRPC**: `buf.build/gen/go/wcygan/todo/connectrpc/go`
- **TypeScript**: `@buf/wcygan_todo.bufbuild_es`

### Updating Dependencies

```bash
# Update Go dependencies to latest generated code
go get buf.build/gen/go/wcygan/todo/connectrpc/go@latest
go get buf.build/gen/go/wcygan/todo/protocolbuffers/go@latest

# Update TypeScript dependencies (in frontend/)
npm install @buf/wcygan_todo.bufbuild_es@latest
```

### Code Generation

```bash
# Generate code from proto definitions
buf generate

# Lint proto files
buf lint

# Format proto files
buf format -w
```

## Server Endpoints

The ConnectRPC server exposes both HTTP and gRPC protocols:

- **Base URL**: `http://localhost:8080`
- **Health Check**: `GET /health`
- **gRPC Service**: `task.v1.TaskService`

### Available Endpoints

| Method | HTTP Path | gRPC Method |
|--------|-----------|-------------|
| POST | `/task.v1.TaskService/CreateTask` | `task.v1.TaskService/CreateTask` |
| POST | `/task.v1.TaskService/GetTask` | `task.v1.TaskService/GetTask` |
| POST | `/task.v1.TaskService/GetAllTasks` | `task.v1.TaskService/GetAllTasks` |
| POST | `/task.v1.TaskService/UpdateTask` | `task.v1.TaskService/UpdateTask` |
| POST | `/task.v1.TaskService/DeleteTask` | `task.v1.TaskService/DeleteTask` |

## Using grpcurl

grpcurl is a command-line tool for interacting with gRPC services.

### Installation

```bash
# macOS
brew install grpcurl

# Go install
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

### Basic Commands

```bash
# List available services (may not work with ConnectRPC)
grpcurl -plaintext localhost:8080 list

# Create a task
grpcurl -plaintext \
  -import-path /path/to/todo/proto \
  -proto task/v1/task.proto \
  -d '{"description": "My new task"}' \
  localhost:8080 task.v1.TaskService/CreateTask

# Get a specific task
grpcurl -plaintext \
  -import-path /path/to/todo/proto \
  -proto task/v1/task.proto \
  -d '{"id": "1"}' \
  localhost:8080 task.v1.TaskService/GetTask

# Get all tasks
grpcurl -plaintext \
  -import-path /path/to/todo/proto \
  -proto task/v1/task.proto \
  -d '{}' \
  localhost:8080 task.v1.TaskService/GetAllTasks

# Update a task
grpcurl -plaintext \
  -import-path /path/to/todo/proto \
  -proto task/v1/task.proto \
  -d '{"id": "1", "description": "Updated task", "completed": true}' \
  localhost:8080 task.v1.TaskService/UpdateTask

# Delete a task
grpcurl -plaintext \
  -import-path /path/to/todo/proto \
  -proto task/v1/task.proto \
  -d '{"id": "1"}' \
  localhost:8080 task.v1.TaskService/DeleteTask
```

### Example Usage

```bash
# Complete workflow example
cd /Users/wcygan/Development/todo

# 1. Create a task
grpcurl -plaintext \
  -import-path proto \
  -proto task/v1/task.proto \
  -d '{"description": "Learn gRPC"}' \
  localhost:8080 task.v1.TaskService/CreateTask

# Output: {"task":{"id":"1","description":"Learn gRPC",...}}

# 2. Get the task
grpcurl -plaintext \
  -import-path proto \
  -proto task/v1/task.proto \
  -d '{"id": "1"}' \
  localhost:8080 task.v1.TaskService/GetTask

# 3. Update the task
grpcurl -plaintext \
  -import-path proto \
  -proto task/v1/task.proto \
  -d '{"id": "1", "description": "Learn gRPC ✓", "completed": true}' \
  localhost:8080 task.v1.TaskService/UpdateTask
```

## Using grpcui

grpcui provides a web-based GUI for testing gRPC services.

### Installation

```bash
# Go install
go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
```

### Launch grpcui

```bash
# Start the web UI
grpcui -plaintext \
  -import-path /path/to/todo/proto \
  -proto task/v1/task.proto \
  localhost:8080

# Opens browser at http://127.0.0.1:<random-port>/
```

### Using the Web Interface

1. **Select Service**: Choose `task.v1.TaskService`
2. **Select Method**: Pick from CreateTask, GetTask, etc.
3. **Fill Request**: Enter JSON data in the request form
4. **Invoke**: Click "Invoke" to send the request
5. **View Response**: See the JSON response and metadata

### Example Requests in grpcui

**CreateTask**:
```json
{
  "description": "Task created via grpcui"
}
```

**GetTask**:
```json
{
  "id": "1"
}
```

**UpdateTask**:
```json
{
  "id": "1",
  "description": "Updated via grpcui",
  "completed": true
}
```

## Using HTTP/JSON API

ConnectRPC also exposes HTTP/JSON endpoints for web compatibility.

### cURL Examples

```bash
# Create task
curl -X POST http://localhost:8080/task.v1.TaskService/CreateTask \
  -H "Content-Type: application/json" \
  -d '{"description": "HTTP API task"}'

# Get task
curl -X POST http://localhost:8080/task.v1.TaskService/GetTask \
  -H "Content-Type: application/json" \
  -d '{"id": "1"}'

# Get all tasks
curl -X POST http://localhost:8080/task.v1.TaskService/GetAllTasks \
  -H "Content-Type: application/json" \
  -d '{}'

# Update task
curl -X POST http://localhost:8080/task.v1.TaskService/UpdateTask \
  -H "Content-Type: application/json" \
  -d '{"id": "1", "description": "Updated task", "completed": true}'

# Delete task
curl -X POST http://localhost:8080/task.v1.TaskService/DeleteTask \
  -H "Content-Type: application/json" \
  -d '{"id": "1"}'
```

## Frontend Integration

The frontend uses generated TypeScript clients:

```typescript
import { taskClient } from "@/lib/client";
import { 
  CreateTaskRequestSchema,
  GetTaskRequestSchema,
  GetAllTasksRequestSchema 
} from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";
import { create } from "@bufbuild/protobuf";

// Create a task
const createRequest = create(CreateTaskRequestSchema, {
  description: "Frontend task"
});
const response = await taskClient.createTask(createRequest);

// Get a task
const getRequest = create(GetTaskRequestSchema, { id: "1" });
const task = await taskClient.getTask(getRequest);
```

## Backend Integration

The backend uses generated Go types:

```go
import (
    taskv1 "buf.build/gen/go/wcygan/todo/protocolbuffers/go/task/v1"
    taskconnect "buf.build/gen/go/wcygan/todo/connectrpc/go/task/v1/taskv1connect"
)

// Handler implementation
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
```

## Troubleshooting

### Common Issues

1. **404 Not Found**: Endpoint not registered
   - Check server logs for registered endpoints
   - Verify handler implements interface correctly
   - Update buf dependencies to latest

2. **Import Errors**: Proto file not found
   - Use correct import path: `-import-path /path/to/todo/proto`
   - Use correct proto file: `-proto task/v1/task.proto`

3. **Connection Refused**: Server not running
   - Check server status: `curl http://localhost:8080/health`
   - Start server: `deno task up` or `skaffold dev`

4. **Unimplemented**: Method not available
   - Verify proto definition includes the method
   - Check handler implements the method
   - Update generated code with `go get @latest`

### Debugging Tips

```bash
# Check server health
curl http://localhost:8080/health

# Test with simple HTTP first
curl -X POST http://localhost:8080/task.v1.TaskService/GetAllTasks \
  -H "Content-Type: application/json" -d '{}'

# Use verbose mode in grpcurl
grpcurl -v -plaintext ...

# Check Go module versions
go list -m buf.build/gen/go/wcygan/todo/connectrpc/go
```

## Development Workflow

1. **Update Proto**: Modify `proto/task/v1/task.proto`
2. **Generate Code**: Run `buf generate` 
3. **Update Dependencies**: 
   - Backend: `go get buf.build/gen/go/wcygan/todo/connectrpc/go@latest`
   - Frontend: `npm install @buf/wcygan_todo.bufbuild_es@latest`
4. **Implement Handlers**: Add/update Go handler methods
5. **Update Frontend**: Use new generated TypeScript types
6. **Test**: Use grpcurl, grpcui, or curl to verify
7. **Deploy**: `skaffold dev` for development

## Resources

- [ConnectRPC Documentation](https://connectrpc.com/)
- [Protocol Buffers Guide](https://protobuf.dev/)
- [Buf Documentation](https://buf.build/docs/)
- [grpcurl GitHub](https://github.com/fullstorydev/grpcurl)
- [grpcui GitHub](https://github.com/fullstorydev/grpcui)