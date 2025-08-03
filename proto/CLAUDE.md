# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a highly available todo list application built with a modern, microservices-oriented architecture using ConnectRPC and Protocol Buffers. The project consists of three main components:

- **Proto definitions** (`/proto/`) - API contracts using Protocol Buffers
- **Go backend** (`/backend/`) - ConnectRPC service implementation  
- **Next.js frontend** (`/frontend/`) - React 19 web application

## Development Commands

### Protocol Buffer Management
```bash
# Generate code from proto files (run from project root)
buf generate

# Lint protocol buffer files
buf lint

# Format protocol buffer files  
buf format -w

# Push schemas to buf.build registry
buf push
```

### Backend Development (Go)
```bash
# Run the server (from backend/ directory)
go run cmd/server/main.go

# Build the application
go build -o server cmd/server/main.go

# Run all tests (unit + integration)
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for specific package
go test ./internal/handler/
```

### Frontend Development (Next.js)
```bash
# Start development server with Turbopack (from frontend/ directory)
npm run dev

# Build for production
npm run build

# Start production server
npm run start

# Run linting
npm run lint
```

## Code Architecture

### API Design Pattern
- **Protocol-first development**: API contracts defined in `.proto` files before implementation
- **ConnectRPC**: Provides both gRPC and REST endpoints from single service definition
- **Code generation**: TypeScript and Go clients/servers generated from proto schemas
- **Buf registry**: Schemas published to `buf.build/wcygan/todo` for versioning and sharing

### Backend Architecture (Go)
- **Clean architecture**: Clear separation between handler, service, and store layers
- **ConnectRPC handlers**: Located in `internal/handler/` - implement the proto service interface
- **Storage layer**: `internal/store/` - currently in-memory, designed for easy database swapping  
- **Main entry point**: `cmd/server/main.go` - sets up HTTP/2 server with h2c for local development
- **Test strategy**: Unit tests for handlers/store, integration tests in `test/integration/`

### Frontend Architecture (Next.js)
- **App Router**: Uses Next.js 15's app directory structure
- **React 19**: Latest React with concurrent features
- **TypeScript**: Strict mode enabled with path aliases (`@/*`)
- **Tailwind CSS 4**: Utility-first styling
- **Generated clients**: ConnectRPC TypeScript client generated from proto files

### Protocol Buffer Schema
The Task service API includes:
- `CreateTask` - Add new tasks
- `GetAllTasks` - Retrieve all tasks  
- `DeleteTask` - Remove tasks by ID

Task entity fields:
- `id` (string) - Unique identifier
- `description` (string) - Task description
- `completed` (bool) - Completion status
- `created_at` (timestamp) - Creation time
- `updated_at` (timestamp) - Last modification time

## Testing Strategy

### Backend Testing
- **Unit tests**: Test individual components (handlers, store) in isolation
- **Integration tests**: Test full HTTP request/response cycle via ConnectRPC
- **Test files pattern**: `*_test.go` files alongside source code
- **Run single test**: `go test -run TestSpecificFunction ./internal/handler/`

### Generated Code Testing
- Protocol buffer schemas automatically validated during `buf generate`
- ConnectRPC client/server compatibility tested through integration tests
- Changes to `.proto` files trigger regeneration of Go and TypeScript clients

## Key Technical Decisions

### Why ConnectRPC over gRPC
- **Web compatibility**: Works directly with browsers and standard HTTP tooling
- **Developer experience**: Simpler than traditional gRPC setup
- **Dual protocol**: Supports both gRPC and REST from same service definition

### Why Protocol Buffers
- **Type safety**: Strongly typed contracts between frontend and backend
- **Code generation**: Eliminates manual client/server boilerplate
- **Versioning**: Buf registry provides schema evolution and compatibility checking
- **Performance**: Efficient binary serialization

### Development Environment
- **HTTP/2 with h2c**: Enables HTTP/2 benefits without TLS complexity in development
- **In-memory storage**: Allows rapid prototyping; easily swappable for production database
- **Turbopack**: Next.js development uses faster Rust-based bundler

## Common Workflows

### Adding New API Endpoints
1. Update `proto/task/v1/task.proto` with new service methods
2. Run `buf generate` to update generated code
3. Implement handler in `backend/internal/handler/`
4. Add corresponding store methods if needed
5. Write unit tests for new functionality
6. Update frontend to use new generated TypeScript client

### Making Schema Changes
1. Modify `.proto` files following protobuf compatibility rules
2. Run `buf lint` to validate changes
3. Run `buf generate` to update generated code
4. Update implementation code to match new schema
5. Run full test suite to ensure compatibility
6. Consider `buf push` to publish new schema version

### Working with the Current Codebase
- Backend is fully functional and tested
- Frontend shows default Next.js page - needs integration with backend API
- Database layer is in-memory - ready for MySQL integration as mentioned in broader project README
- Kubernetes deployment configuration is planned but not yet implemented