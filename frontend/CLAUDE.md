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

This is a **Next.js 15 todo list application** built according to the `design.md` specification. It is a UI-only prototype with mock data to demonstrate a modern, clean, and efficient task management interface. The application logic is handled entirely on the client-side, with state persistence managed via `localStorage`.

**Status:** This is a new project foundation. The goal is to build out the features described in `design.md`.

## Technology Stack

-   **Framework:** Next.js 15 with App Router
-   **Language:** TypeScript
-   **Styling:** Tailwind CSS with PostCSS
-   **Component Library:** shadcn/ui
-   **State Management:** React Context + useReducer
-   **Forms:** React Hook Form + Zod

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

### Page Routes & Information Architecture
```
/              # Main view: Active and overdue tasks.
/completed     # View for all completed tasks.
/settings      # (Future Scope) User preferences.

# Modals (Dialogs)
- Add/Edit Task Modal: For creating and editing tasks.
```

### Proposed Directory Structure
```
/frontend/
├── src/
│   ├── app/
│   │   ├── layout.tsx         # Global shell with TodoProvider, Toaster
│   │   ├── page.tsx           # Main active task list view
│   │   └── completed/page.tsx # Completed tasks view
│   ├── components/
│   │   ├── ui/                # shadcn/ui components
│   │   ├── tasks/             # TodoItemRow, AddTaskForm, etc.
│   │   └── layout/            # Header, etc.
│   ├── lib/
│   │   ├── mockApi.ts         # Fake API for tasks
│   │   ├── utils.ts           # cn() helper
│   │   └── validation/        # Zod schemas for tasks
│   ├── store/
│   │   ├── todo.tsx           # TodoProvider (Context + reducer)
│   │   └── todoReducer.ts     # Task state logic
│   ├── types/
│   │   └── index.ts           # Core types (Task, Priority, etc.)
│   └── hooks/
│       └── useTodos.ts        # Hook for interacting with todo state
└── public/
    └── icons/                 # Placeholder for any custom icons
```

## Core Data Models

The primary types for the application are:
```typescript
// types/index.ts
type Priority = 'low' | 'medium' | 'high' | 'none';

type Task = {
  id: string;
  title: string;
  isCompleted: boolean;
  priority: Priority;
  dueDate?: string; // ISO date string
  notes?: string;
  createdAt: string; // ISO date string
  completedAt?: string; // ISO date string
};
```

## State Management

### Todo State Architecture
-   **`TodoProvider`**: Manages all task state using React Context + `useReducer`.
-   **Actions**: `ADD_TASK`, `TOGGLE_TASK`, `UPDATE_TASK`, `DELETE_TASK`, `SET_FILTER`.
-   **Persistence**: The entire task list is synchronized with `localStorage` (with throttling).
-   **`useTodos()` hook**: Provides typed access to the task list, filters, and dispatcher functions.

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

## Common Development Tasks

### Creating a New Task
1.  Invoke the `useTodos().add` method.
2.  The `ui-design-enforcer` should ensure the "Add Task" button triggers a `Dialog` with the `AddTaskForm`.
3.  The `task-manager` ensures the new task is added to state and persisted to `localStorage`.

### Modifying Task State (e.g., Toggling Completion)
1.  Call the `toggleTask` action from the `useTodos` hook.
2.  The `ui-design-enforcer` is responsible for ensuring the UI correctly reflects the `isCompleted` state (e.g., opacity, line-through text).
3.  The `task-manager` handles the state change in the reducer.

### Updating Styles
1.  **ALWAYS** check `design.md` first.
2.  Use only approved color tokens and spacing.
3.  Test all responsive breakpoints and interactive states (hover, focus, completed).

## Common Issues & Solutions

### Task state not persisting on refresh
-   Verify the `TodoProvider` wraps the entire application in `layout.tsx`.
-   Check the browser's DevTools to ensure data is being written to `localStorage` under the correct key.
-   Ensure the hydration logic in `TodoProvider` is correctly parsing the stored JSON.

### UI not updating after a state change
-   This is likely a missing state dependency or incorrect reducer logic.
-   The `task-manager` should review the reducer to ensure a new state object is returned.
-   The `type-architect` should check component dependencies in hooks (`useEffect`, `useMemo`).

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