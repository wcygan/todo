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

### Directory Structure
```
/
├── proto/           # Protocol Buffer definitions (buf.build/wcygan/todo)
├── backend/         # Go ConnectRPC service with comprehensive CLAUDE.md
├── frontend/        # Next.js 15 App Router with specialized agent architecture
├── k8s/             # Kubernetes manifests and deployment configurations
├── docs/            # Technical documentation including k8s-development.md
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

1. **ConnectRPC over gRPC**: Modern HTTP/2 RPC with better web compatibility
2. **Protocol Buffer Schema Registry**: Centralized schema management via buf.build
3. **Specialized Frontend Agents**: Multi-agent coordination for UI consistency
4. **MySQL Database**: Production-ready persistence with MySQL Operator for Kubernetes
5. **Type Safety**: End-to-end TypeScript + Go with generated protobuf types
6. **Accessibility First**: WCAG AA compliance built into design system

## Production Considerations

- **Kubernetes Infrastructure**: Fully configured with Tilt for development-to-production parity
- **Database**: MySQL with automatic backups, scaling, and high availability via MySQL Operator
- **Authentication**: No auth layer currently implemented (planned enhancement)
- **Monitoring**: Basic health checks implemented, comprehensive observability planned
- **State Persistence**: Frontend state is ephemeral (no localStorage)
- **Scaling**: Horizontal pod autoscaling and multi-replica deployments configured

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
- **Component Documentation**: Each major component (frontend/backend) has detailed CLAUDE.md files with specific requirements
- **Frontend Agents**: Mandatory specialized agent consultation before any frontend changes
- **Design Compliance**: All UI must strictly follow the design.md specification
- **Database**: MySQL Operator manages production-ready database with automatic operations
- **Testing**: Comprehensive test coverage maintained across all layers
- **Protocol Buffers**: Schemas are the source of truth for API contracts
- **Documentation**: See `docs/k8s-development.md` for complete Kubernetes setup and troubleshooting

## Buf Schema Registry Workflow

- Instead of doing `buf generate`, modify a proto file in @proto/ then push to origin/main
- Pushing to origin/main will trigger SDK generation on the Buf Schema Registry
- After pushing, update to the latest version of the dependency
- Specifically, after modifying proto schemas, you will need to do updates like this:
  - For the frontend: `bun add @buf/wcygan_todo.bufbuild_es@latest` and `bun add @buf/wcygan_todo.connectrpc_query-es@latest`
  - For the Backend: `go get buf.build/gen/go/wcygan/todo/protocolbuffers/go@latest` and `go get buf.build/gen/go/wcygan/todo/connectrpc/go@latest`
