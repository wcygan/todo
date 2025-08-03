# CLAUDE.md

This file provides guidance to Claude Code when working on this repository. [docs.anthropic.com](https://docs.anthropic.com/en/docs/claude-code/overview)

## ⚠️ CRITICAL: Specialized Sub-Agents Required

**BEFORE making any changes, you MUST leverage the specialized sub-agents.**

### Three Specialized Agents Available
1.  **`ui-design-enforcer`** - UI/UX design enforcer (red) - Ensures strict `design.md` compliance.
2.  **`task-manager`** - Todo logic specialist (blue) - Handles state, CRUD operations, filtering, and core app logic.
3.  **`type-architect`** - TypeScript architecture specialist (green) - Ensures type safety & Next.js patterns.

### Mandatory Agent Consultation Rules
-   **ALL UI work**: Consult `ui-design-enforcer` FIRST before any styling or component changes.
-   **ANY task feature**: Use `task-manager` for adding, editing, completing, filtering, or managing task state.
-   **TypeScript/architecture**: Engage `type-architect` for component structure, types, Next.js patterns, or fixing compilation errors.
-   **Complex features**: Use multiple agents in coordination (e.g., design → logic → types).

## Project Overview

This is a **Next.js 15 todo list application** with full-stack ConnectRPC integration. The frontend communicates with a Go ConnectRPC backend using Protocol Buffer schemas published to buf.build/wcygan/todo. The application features modern data fetching with TanStack Query for optimal caching and real-time updates.

**Status:** Core functionality complete with ConnectRPC integration. Ready for UI enhancement following `design.md` specification.

## Technology Stack

-   **Framework:** Next.js 15 with App Router
-   **Language:** TypeScript
-   **Styling:** Tailwind CSS with PostCSS
-   **Component Library:** shadcn/ui (planned)
-   **Backend Integration:** ConnectRPC with Protocol Buffers
-   **Data Fetching:** TanStack Query v5 for caching and state management
-   **API Client:** @connectrpc/connect-web with buf.build generated types
-   **Forms:** React Hook Form + Zod (planned)

## Development Commands

**Working Directory:** All commands should be run from the `/frontend/` subdirectory.

```bash
# Start development server with Turbopack
bun dev

# Create a production build
bun run build

# Lint the code
bun run lint

# Check for TypeScript errors
bunx tsc --noEmit
```

## Architecture & Implementation

### ConnectRPC Integration

The frontend integrates with a Go ConnectRPC backend using modern Protocol Buffer schemas:

**Backend Integration:**
- **API Endpoint:** `http://localhost:8080`
- **Schema Registry:** buf.build/wcygan/todo
- **Generated Types:** `@buf/wcygan_todo.bufbuild_es`
- **Transport:** HTTP/2 with JSON support via ConnectRPC

**Available API Methods:**
- `CreateTask(description: string) → Task`
- `GetAllTasks() → Task[]`
- `DeleteTask(id: string) → {success: boolean, message: string}`

### Data Flow Architecture

```
User Action → TanStack Query → ConnectRPC Client → Go Backend
     ↓              ↓                ↓              ↓
  UI Update ← Cache Update ← JSON Response ← Protocol Buffer
```

**Key Benefits:**
- **Type Safety:** End-to-end TypeScript types from protobuf schemas
- **Caching:** Automatic caching and background refetching with TanStack Query
- **Real-time:** Optimistic updates with automatic cache invalidation
- **Error Handling:** Built-in retry mechanisms and error boundaries

### Current Directory Structure
```
/frontend/
├── src/
│   ├── app/
│   │   ├── layout.tsx         # Global shell with QueryProvider
│   │   ├── page.tsx           # Main todo app with TanStack Query
│   │   └── test/page.tsx      # API testing interface
│   ├── components/
│   │   └── providers/
│   │       └── query-provider.tsx  # TanStack Query setup
│   ├── lib/
│   │   ├── client.ts          # ConnectRPC client configuration
│   │   └── query-client.ts    # TanStack Query client setup
│   └── (shadcn/ui and other components planned)
└── node_modules/
    └── @buf/wcygan_todo.bufbuild_es/  # Generated protobuf types
```

### Page Routes & Information Architecture
```
/              # Main todo app with ConnectRPC integration
/test          # API testing interface for all endpoints
/completed     # (Future) View for completed tasks
/settings      # (Future) User preferences
```

## Core Data Models

### Protocol Buffer Schema

The application uses generated TypeScript types from Protocol Buffer definitions:

```typescript
// Generated from buf.build/wcygan/todo
import { Task, CreateTaskRequest, GetAllTasksRequest, DeleteTaskRequest } from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";

// Task structure from protobuf schema:
type Task = {
  id: string;
  description: string;
  completed: boolean;
  createdAt?: Timestamp;  // google.protobuf.Timestamp
  updatedAt?: Timestamp;  // google.protobuf.Timestamp
};
```

**Note:** The current backend schema is simplified compared to the frontend design specification. Future enhancements will add:
- `priority` field (low/medium/high/none)
- `dueDate` field for scheduling
- `notes` field for additional context
- `completedAt` timestamp

## State Management

### TanStack Query Architecture

**Query Configuration:**
```typescript
// lib/query-client.ts
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5000,        // Data fresh for 5 seconds
      gcTime: 10 * 60 * 1000, // Cache for 10 minutes
    },
  },
});
```

**Data Fetching Pattern:**
```typescript
// Query for getting tasks
const { data, isLoading, error } = useQuery({
  queryKey: ["tasks"],
  queryFn: async () => {
    const request = create(GetAllTasksRequestSchema, {});
    return await taskClient.getAllTasks(request);
  },
});

// Mutations for CRUD operations
const createTaskMutation = useMutation({
  mutationFn: async (description: string) => {
    const request = create(CreateTaskRequestSchema, { description });
    return await taskClient.createTask(request);
  },
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ["tasks"] });
  },
});
```

## `design.md` Enforcement

**⚠️ MANDATORY**: All styling and component implementation MUST follow the `design.md` specification exactly. The `ui-design-enforcer` agent is the enforcer of these rules.

### Key `design.md` Integration Rules
1.  **Before ANY UI work**: Read `design.md` to understand color tokens, spacing, and component patterns.
2.  **Task Rows**: Implement `TodoItemRow` exactly as specified, including hover states and completed states (`line-through`).
3.  **Color Usage**: Use status colors (`bg-red-100`, `text-slate-400`) and priority badges as defined. **No arbitrary colors.**
4.  **Spacing**: Adhere strictly to the 8pt grid system (`py-8`, `gap-4`, `p-6`).
5.  **Accessibility**: Ensure all interactive elements meet the WCAG AA requirements outlined in `design.md`, including focus rings and ARIA labels for icon buttons.

## Form Handling

### Add/Edit Task Form
**Stack:** React Hook Form + Zod + shadcn/ui Form components, presented inside a `Dialog`.

**Validation Schema (`/lib/validation/task.ts`):**
```typescript
import { z } from "zod";
export const taskSchema = z.object({
  title: z.string().min(3, "Task title must be at least 3 characters."),
  priority: z.enum(['none', 'low', 'medium', 'high']),
  dueDate: z.string().optional(),
  notes: z.string().optional(),
});
```

## Agent Coordination Rules & Quality Gates

### Agent-Integrated Workflow
-   **New UI Component (e.g., `PriorityBadge`)**:
    1.  `ui-design-enforcer`: Create the component shell with correct styling from `design.md`.
    2.  `type-architect`: Add TypeScript props and ensure type safety.
-   **New Feature (e.g., Filtering)**:
    1.  `task-manager`: Implement the state logic in the reducer and context.
    2.  `ui-design-enforcer`: Build the UI controls (e.g., buttons, dropdown) for the filter.
    3.  `type-architect`: Connect the UI to the state logic, ensuring types match.

### Multi-Agent Definition of Done
**EVERY feature must pass ALL agent validations:**
1.  **`ui-design-enforcer`**: UI matches `design.md` exactly.
2.  **`task-manager`**: Core logic is sound and state updates correctly.
3.  **`type-architect`**: TypeScript compiles without errors and architecture is clean.

## Current Functionality (✅ Tested & Working)

### Implemented Features
- **✅ Task Display:** Lists all tasks from backend with real-time updates
- **✅ Create Task:** Add new tasks via form input with auto-refresh
- **✅ Delete Task:** Remove tasks with immediate UI update
- **✅ Error Handling:** Comprehensive error states with retry functionality
- **✅ Loading States:** Proper loading indicators for all operations
- **✅ API Testing:** Dedicated `/test` page for manual API verification

### ConnectRPC Integration Status
- **✅ Client Setup:** ConnectRPC client configured for HTTP/2 + JSON
- **✅ Type Safety:** Generated TypeScript types from buf.build registry
- **✅ CORS Support:** Backend configured for web client communication
- **✅ Cache Management:** TanStack Query handles automatic invalidation
- **✅ Optimistic Updates:** UI responds immediately to user actions

### Testing Verified With Puppeteer
- **✅ Page Load:** App loads correctly with task list
- **✅ Create Flow:** Form submission → API call → UI update → form reset
- **✅ Delete Flow:** Button click → API call → task removal → cache refresh
- **✅ Error Recovery:** Network failures show error states with retry options
- **✅ Navigation:** Test page accessible and functional

## Common Development Tasks

### Creating a New Task (Current Implementation)
1. User types in input field and submits form
2. `createTaskMutation` triggers with description
3. ConnectRPC client sends `CreateTask` request to backend
4. On success, TanStack Query invalidates `["tasks"]` cache
5. UI automatically refetches and displays new task

### Deleting a Task (Current Implementation)
1. User clicks delete button on task
2. `deleteTaskMutation` triggers with task ID
3. ConnectRPC client sends `DeleteTask` request to backend
4. On success, TanStack Query invalidates `["tasks"]` cache
5. UI automatically refetches and removes deleted task

### Adding New Features
1. **Backend First:** Update protobuf schema and regenerate types
2. **Frontend Integration:** Update client calls and UI components
3. **Query Updates:** Modify TanStack Query keys and invalidation logic
4. **UI Compliance:** Follow `design.md` for all styling changes

## Common Issues & Solutions

### ConnectRPC Client Errors
- **Import Issues:** Use `createClient` instead of `createPromiseClient` for ConnectRPC v2
- **CORS Errors:** Ensure backend CORS configuration includes frontend origin
- **Type Errors:** Regenerate types from buf.build if protobuf schema changes

### TanStack Query Issues
- **Stale Data:** Check query keys match between queries and mutations
- **Cache Not Updating:** Ensure `queryClient.invalidateQueries()` called in mutation `onSuccess`
- **Loading States:** Verify `isLoading` and `isPending` states are handled in UI

### Protobuf Type Issues
- **Timestamp Conversion:** Use `new Date(timestamp.seconds * 1000)` for display
- **Schema Mismatches:** Verify generated types match backend implementation
- **Import Paths:** Use `.js` extension in imports for proper ESM compatibility

## Puppeteer MCP for Development & Debugging

Use the Puppeteer MCP tool to automate and visually verify UI changes.

#### Example 1: Add a New Task and Verify
```javascript
// Navigate to the app
mcp__puppeteer__puppeteer_navigate({ url: "http://localhost:3000" })

// Click the "Add Task" button
mcp__puppeteer__puppeteer_click({ selector: 'button:contains("Add Task")' })

// Wait for the dialog to appear and fill the form
mcp__puppeteer__puppeteer_fill({ selector: 'input[name="title"]', value: "My new test task" })
mcp__puppeteer__puppeteer_click({ selector: 'button[type="submit"]' })

// Take a screenshot to verify the task was added
mcp__puppeteer__puppeteer_screenshot({
  name: "after-task-added",
  width: 1200,
  height: 800
})
```

#### Example 2: Complete a Task and Verify Style Change
```javascript
// Find the checkbox for the first task and click it
mcp__puppeteer__puppeteer_click({ selector: '[role="checkbox"]:first' })

// Take screenshot to verify the line-through style
mcp__puppeteer__puppeteer_screenshot({
  name: "after-task-completed",
  width: 1200,
  height: 800
})

// Bonus: Verify the style with JS
mcp__puppeteer__puppeteer_evaluate({
  script: "document.querySelector('label:first').style.textDecoration"
})
```