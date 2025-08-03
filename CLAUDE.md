# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A modern, highly available todo list application with full-stack architecture using Protocol Buffers for type-safe communication between services.

**Tech Stack:**
- **Frontend**: Next.js 15 + React 19 + TypeScript + Tailwind CSS 4
- **Backend**: Go 1.24.2 + ConnectRPC + Protocol Buffers
- **Database**: MySQL 9.0 with MySQL Operator for Kubernetes
- **Infrastructure**: Kubernetes + Tilt for development
- **Schema Management**: Buf Schema Registry: buf.build/wcygan/todo

## Architecture

### Backend for Frontend (BFF) Pattern
This application uses the **Backend for Frontend (BFF)** pattern where Next.js serves as both the UI and API proxy layer:

```
Internet → Ingress → Next.js (Frontend + API Routes) → Backend Service (Internal)
                           ↓
                     Browser/Client
```

**Key Benefits:**
- **Security**: Backend API never exposed to internet
- **Simplicity**: No CORS configuration needed
- **Type Safety**: Shared types between frontend and API proxy
- **Single Deployment**: Only Next.js needs external exposure

### Directory Structure
```
/
├── proto/           # Protocol Buffer definitions (buf.build/wcygan/todo)
├── backend/         # Go ConnectRPC service (internal, ClusterIP only)
├── frontend/        # Next.js 15 with BFF API routes
│   └── src/app/api/ # API proxy routes to backend
├── k8s/             # Kubernetes manifests and deployment configurations
├── docs/            # Technical documentation including production-architecture-decision.md
├── buf.gen.yaml     # Buf code generation configuration
├── Tiltfile         # Tilt development workflow configuration
└── .github/         # CI/CD for automated protobuf registry publishing
```

### Core Data Model
```typescript
type Task = {
  id: string;
  title: string;
  isCompleted: boolean;
  priority: 'low' | 'medium' | 'high' | 'none';
  dueDate?: string;
  notes?: string;
  createdAt: string;
  completedAt?: string;
};
```

### API Design
ConnectRPC service with four endpoints:
- `CreateTask` - Create new task
- `GetAllTasks` - Retrieve all tasks
- `DeleteTask` - Remove task by ID
- `UpdateTask` - Update task properties (including completion status)

## Development Commands

### Kubernetes Development (from project root)
```bash
# Start full development environment
tilt up                 # Start all services with hot reload and live updates
tilt down               # Stop all services and clean up resources

# Development operations
tilt trigger protobuf-gen    # Regenerate protocol buffer code
tilt trigger backend-test    # Run backend tests
tilt trigger frontend-test   # Run frontend tests
tilt trigger db-migrate      # Run database migrations
```

### Tilt CLI Debugging & Monitoring
```bash
# Inspect deployment status
tilt get all                     # List all Tilt resources and their status
tilt get resources               # Get all resources with detailed status
tilt describe <resource>         # Detailed info about specific resource (frontend, backend, etc.)

# Monitor logs
tilt logs <resource>             # Stream logs from specific resource
tilt logs frontend --tail=50     # Last 50 lines from frontend
tilt logs backend -f             # Follow backend logs in real-time

# Debugging & diagnostics
tilt doctor                      # Check Tilt environment health and configuration
tilt dump engine                 # Dump internal Tilt state for debugging
tilt wait --for=condition=Ready resource/<name>  # Wait for resource to become ready

# Resource management
tilt enable <resource>           # Enable a disabled resource
tilt disable <resource>          # Disable a resource temporarily
tilt trigger <resource>          # Manually trigger resource update

# Examples for this project:
tilt describe frontend           # Check frontend deployment details
tilt logs backend --tail=20      # Recent backend logs
tilt get resources | grep -E "(frontend|backend)"  # Filter for app resources
```

### Local Development (alternative to Kubernetes)

#### Frontend (from `/frontend/`)
```bash
bun dev                  # Development server with Turbopack
bun run build           # Production build
bun run lint            # ESLint checking
bunx tsc --noEmit       # TypeScript type checking
```

#### Backend (from `/backend/`)
```bash
air                     # Hot reload development server
go run ./cmd/server     # Direct server execution
go test ./...           # Run all tests (unit + integration)
go build -o server ./cmd/server  # Production build
```

#### Protocol Buffers (from project root)
```bash
buf generate            # Generate Go/TypeScript code from protobuf
buf push                # Publish schemas to buf.build registry
```

## Architecture Guidelines

### Frontend Architecture with BFF

**API Proxy Implementation**: The frontend includes API routes that proxy all backend calls:
- **Client Code**: Always uses relative `/api/*` paths
- **API Routes**: Located in `/frontend/src/app/api/[...path]/route.ts`
- **Backend URL**: Configured via `BACKEND_URL` environment variable (server-side only)
- **No CORS needed**: Same-origin requests from browser to Next.js

**Example Flow**:
1. Browser calls `/api/task.v1.TaskService/GetAllTasks`
2. Next.js API route receives request
3. API route forwards to `http://backend-service:8080/task.v1.TaskService/GetAllTasks`
4. Response flows back through Next.js to browser

### Frontend Requirements
**CRITICAL**: The frontend has specialized agent architecture requirements defined in `/frontend/CLAUDE.md`. Before making ANY frontend changes, you must:

1. Consult the three mandatory specialized agents:
   - `ui-design-enforcer` - Ensures design.md compliance
   - `task-manager` - Handles state and business logic
   - `type-architect` - Maintains TypeScript type safety

2. Follow the comprehensive UI specification in `/frontend/design.md`:
   - 8pt grid spacing system
   - WCAG AA accessibility compliance
   - shadcn/ui component library only
   - Emerald color palette for primary actions

3. Use required libraries:
   - React Hook Form + Zod for form validation
   - React Context + useReducer for state management
   - shadcn/ui for all UI components

### Backend Architecture
- **Clean Architecture**: Handler → Store layer separation
- **Thread Safety**: All store operations use sync.RWMutex
- **Error Handling**: Custom error types with proper ConnectRPC responses
- **Testing**: Comprehensive unit + integration test coverage required

### Protocol-First Development
- All API changes must start with protobuf schema updates
- Use `buf generate` to regenerate client/server code
- Published schemas available at buf.build/wcygan/todo
- Automated CI pushes protobuf changes to registry

## Testing Strategy

### Backend Testing
```bash
go test ./internal/handler/...    # Unit tests for handlers
go test ./internal/store/...      # Unit tests for data layer
go test ./test/integration/...    # Full HTTP integration tests
```

### Test Patterns
- Table-driven tests for Go handlers and store
- Concurrent operation testing for thread safety
- Mock store implementations for isolated testing
- Comprehensive error case coverage

## Key Architectural Decisions

1. **Backend for Frontend (BFF) Pattern**: Next.js API routes proxy to internal backend service
2. **ConnectRPC over gRPC**: Modern HTTP/2 RPC with better web compatibility
3. **Protocol Buffer Schema Registry**: Centralized schema management via buf.build
4. **Specialized Frontend Agents**: Multi-agent coordination for UI consistency
5. **MySQL Database**: Production-ready persistence with simple StatefulSet deployment
6. **Type Safety**: End-to-end TypeScript + Go with generated protobuf types
7. **Accessibility First**: WCAG AA compliance built into design system
8. **Security by Design**: Backend never exposed to internet, all traffic through Next.js

## Production Considerations

- **Kubernetes Infrastructure**: Fully configured with Tilt for development-to-production parity
- **Service Exposure**: Only Next.js frontend exposed via LoadBalancer/Ingress, backend remains ClusterIP
- **Database**: MySQL StatefulSet with persistent volumes for data durability
- **Authentication**: No auth layer currently implemented (planned enhancement)
- **Monitoring**: Basic health checks implemented, comprehensive observability planned
- **State Persistence**: Frontend state is ephemeral (no localStorage)
- **Scaling**: Horizontal pod autoscaling and multi-replica deployments configured
- **API Security**: All backend calls proxied through Next.js API routes for enhanced security

## Development Workflow

### Kubernetes-First Development (Recommended)
1. **Environment Setup**: `tilt up` → starts all services with live reload
2. **Protocol Changes**: Update `.proto` files → Tilt auto-regenerates code → services restart
3. **Backend Development**: Edit Go code → Tilt live-syncs → Air hot reload
4. **Frontend Development**: Edit React/Next.js → Tilt live-syncs → HMR refresh
5. **Database Changes**: Run migrations via `tilt trigger db-migrate`
6. **Testing**: Use `tilt trigger` commands for comprehensive testing

### Alternative Local Development
1. **Protocol Changes**: Update `.proto` files → `buf generate` → test endpoints
2. **Frontend Features**: Consult specialized agents → implement per design.md → accessibility testing
3. **Backend Features**: Write tests first → implement handler/store layers → integration tests
4. **Full Stack**: Test backend endpoint → implement frontend UI → end-to-end validation

## Important Notes

- **Kubernetes-First**: Use `tilt up` for development - provides production-like environment with fast iteration
- **BFF Architecture**: Frontend API routes proxy all backend calls - no direct backend exposure
- **Component Documentation**: Each major component (frontend/backend) has detailed CLAUDE.md files with specific requirements
- **Frontend Agents**: Mandatory specialized agent consultation before any frontend changes
- **Design Compliance**: All UI must strictly follow the design.md specification
- **Database**: MySQL StatefulSet provides simple, reliable database deployment
- **Testing**: Comprehensive test coverage maintained across all layers
- **Protocol Buffers**: Schemas are the source of truth for API contracts
- **Documentation**: See `docs/production-architecture-decision.md` for BFF pattern details and `docs/k8s-development.md` for Kubernetes setup

## Common Issues & Solutions

### Frontend Hanging on Load
**Problem**: Browser can't resolve Kubernetes service names like `backend-service`
**Solution**: Use BFF pattern - frontend calls `/api/*` which Next.js proxies to backend

### MySQL Operator Issues
**Problem**: MySQL Operator fails to initialize cluster with authentication errors
**Solution**: Use simple MySQL StatefulSet instead - see `/k8s/development/mysql-simple.yaml`

### CORS Errors
**Problem**: Browser blocks cross-origin requests to backend
**Solution**: BFF pattern eliminates CORS - all requests are same-origin to Next.js

### Environment-Specific URLs
**Problem**: Different API URLs for development vs production
**Solution**: Frontend always uses `/api/*`, Next.js handles backend URL via environment variable

## Buf Schema Registry Workflow

- Instead of doing `buf generate`, modify a proto file in @proto/ then push to origin/main
- Pushing to origin/main will trigger SDK generation on the Buf Schema Registry
- After pushing, update to the latest version of the dependency
- Specifically, after modifying proto schemas, you will need to do updates like this:
  - For the frontend: `bun add @buf/wcygan_todo.bufbuild_es@latest` and `bun add @buf/wcygan_todo.connectrpc_query-es@latest`
  - For the Backend: `go get buf.build/gen/go/wcygan/todo/protocolbuffers/go@latest` and `go get buf.build/gen/go/wcygan/todo/connectrpc/go@latest`
