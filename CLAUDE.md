# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

**Primary workflow (use Deno task runner):**
- `deno task up` - Start all services with Docker Compose
- `deno task down` - Stop all services  
- `deno task dev` - Build and start services for development
- `deno task logs` - Follow service logs
- `deno task restart` - Restart all services
- `deno task clean` - Stop services and clean Docker system

**Testing:**
- `go test ./...` - Run all Go backend tests
- `go test -v ./internal/...` - Run backend unit tests with verbose output
- `go test -v ./test/integration/...` - Run backend integration tests
- `npm run lint` - Run frontend linting (from frontend/ directory)

**Development with Kubernetes:**
- `skaffold dev --profile=dev` - Hot-reload development with port forwarding
- `skaffold run` - Build and deploy once
- `skaffold delete` - Clean up deployed resources

**Services run on:**
- Frontend: http://localhost:3000
- Backend: http://localhost:8080
- Backend health: http://localhost:8080/health

## Architecture Overview

**Technology Stack:**
- **Backend**: Go 1.24.2 with ConnectRPC (modern gRPC alternative)
- **Frontend**: Next.js 15.4.5 with React 19, TypeScript, TailwindCSS
- **API**: Protocol Buffers with Buf code generation
- **Storage**: Currently in-memory (designed for MySQL integration)
- **Deployment**: Kubernetes with Skaffold, Docker containers

**Protocol-First Design:**
- Single API contract defined in `proto/task/v1/task.proto`
- Four RPC methods: `CreateTask`, `GetAllTasks`, `UpdateTask`, `DeleteTask`
- Generated clients for both Go backend and TypeScript frontend
- Full type safety from Protocol Buffers through both services

## Backend Architecture (Go + ConnectRPC)

**Clean Architecture Layers:**
```
Handler Layer    → ConnectRPC endpoints (internal/handler/)
Service Layer    → Business logic (internal/service/)
Store Layer      → Data access (internal/store/)
```

**Key Components:**
- **Entry Point**: `backend/cmd/server/main.go`
- **Handlers**: ConnectRPC endpoint implementations
- **Services**: Business logic with validation
- **Store**: Thread-safe in-memory storage (interface-ready for database)
- **Middleware**: Request logging, timeouts, CORS
- **Config**: Environment-driven configuration in `internal/config/`

**Error Handling:**
- Custom error types with codes: `NOT_FOUND`, `VALIDATION_ERROR`, `INTERNAL_ERROR`, `TIMEOUT`
- Automatic conversion to ConnectRPC status codes
- Structured error context preservation

**Testing Strategy:**
- Unit tests for all layers (`*_test.go` files)
- Integration tests in `test/integration/`
- Test utilities in `test/testutil/`

## Frontend Architecture (Next.js + ConnectRPC)

**Modern React Patterns:**
- Next.js App Router with TypeScript
- TanStack Query for server state management
- `@connectrpc/connect-web` for API communication
- React Hook Form for form handling
- Zod for runtime validation

**Component Structure:**
- `app/` - Next.js App Router pages
- `components/` - Reusable UI components with shadcn/ui
- `lib/` - Client configuration and utilities
- `types/` - TypeScript type definitions

## Database Integration

**Current State**: In-memory storage with thread-safe operations
**Planned**: MySQL integration (mentioned in README)
**Pattern**: Repository interface in `internal/store/interface.go` ready for database implementation

## Deployment & Infrastructure

**Local Development:**
- Docker Compose setup with proper service dependencies
- Skaffold with hot-reload for both services
- File sync for Go and TypeScript files

**Kubernetes Deployment:**
- Production-ready manifests in `k8s/`
- Security-hardened containers (non-root, read-only filesystems)
- Proper resource limits and health checks
- Dedicated `todo-app` namespace

**Container Strategy:**
- Multi-stage Docker builds
- Distroless base images for security
- Layer optimization for build caching

## Development Patterns

**Code Generation:**
- Protocol Buffers generate Go and TypeScript clients
- Buf for linting and code generation
- Generated code lives in respective language directories

**Configuration Management:**
- Environment-specific configs in `backend/configs/`
- Environment variables for runtime configuration
- Development vs production profiles

**API Communication:**
- HTTP/2 with h2c for development
- ConnectRPC provides better HTTP compatibility than traditional gRPC
- Full CORS support for web clients

## Key Files to Understand

**Core API Contract:**
- `proto/task/v1/task.proto` - Single source of truth for API

**Backend Entry Points:**
- `backend/cmd/server/main.go` - Server startup and configuration
- `backend/internal/handler/task.go` - ConnectRPC endpoint implementations

**Frontend Entry Points:**
- `frontend/src/app/page.tsx` - Main application page
- `frontend/src/lib/client.ts` - ConnectRPC client configuration

**Infrastructure:**
- `skaffold.yaml` - Development and deployment workflows
- `docker-compose.yml` - Local development environment
- `k8s/` - Production Kubernetes manifests

This is a modern, well-architected full-stack application demonstrating current best practices for Go backend development, React frontend development, and cloud-native deployment patterns.