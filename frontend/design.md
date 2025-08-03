# Design System & UI Specification (Todo List App)

This document is the single source of truth for all UI/UX decisions for this Next.js todo list application.

## Information Architecture

### Page & View Flow
1.  **Main List View** (`/`) - The primary view showing active and overdue tasks. Contains controls for adding, filtering, and sorting tasks.
2.  **Completed View** (`/completed`) - A dedicated view showing all completed tasks, grouped by date.
3.  **Task Edit/Detail Modal** - A modal overlay for creating a new task or editing an existing one in detail.
4.  **Settings** (`/settings`) - (Future scope) Page for user preferences, themes, etc.

### Core User Flow
User opens app -> sees their task list -> clicks "Add Task" -> fills out the form in a modal -> new task appears in the list -> user clicks a checkbox to complete a task -> task becomes visually completed and moves to the Completed View.

## Color Palette & Design Tokens

This application uses a focused and actionable color palette.

### Primary Colors
-   **Background**: `bg-white`, `bg-stone-50` (for side panels or grouped sections).
-   **Text Primary**: `text-slate-900`
-   **Text Secondary**: `text-slate-600`
-   **Accent/CTA**: `emerald-500`, `hover:bg-emerald-600` (for "Add Task" and other positive actions).
-   **Borders**: `border-slate-200` (1px).
-   **Shadows**: `shadow-sm`, `shadow-md`.

### Status & Priority Colors
-   **High Priority**: `bg-red-100`, `text-red-800`
-   **Medium Priority**: `bg-yellow-100`, `text-yellow-800`
-   **Low Priority**: `bg-blue-100`, `text-blue-800`
-   **Completed**: `text-slate-400`, `line-through` decoration.
-   **Overdue**: `text-red-600` (for due dates).

## Spacing System (8pt Grid)

A consistent 8pt grid ensures clean, rhythmic layouts.

### Section Spacing
-   **Page Header Padding**: `py-8` (32px)
-   **Component Groups**: `gap-4` or `gap-6` (16px / 24px)
-   **Task List Spacing**: `space-y-2` (8px between task items)
-   **Modal Padding**: `p-6` (24px)
-   **Button Padding**: `px-4 py-2` (16px/8px)

### Responsive Breakpoints
-   **Mobile**: Default (≥360px)
-   **Tablet**: `sm:` (≥640px)
-   **Desktop**: `lg:` (≥1024px)

## Component Standards

### Buttons (shadcn/ui only)
```tsx
// Primary CTA
<Button>Add Task</Button>

// Secondary / Cancel
<Button variant="outline">Cancel</Button>

// Destructive
<Button variant="destructive">Delete Task</Button>

// Icon-only (e.g., for edit/delete on rows)
<Button variant="ghost" size="icon">
  <Trash2 className="h-4 w-4" />
</Button>
```

### Todo Item Row
The core component for displaying a single task.

```tsx
// Structure: Checkbox -> Title & Details -> Actions
<div class="flex items-center p-3 hover:bg-stone-50 rounded-lg">
  <Checkbox id="task-1" />
  <div class="flex-1 mx-4">
    <label htmlFor="task-1" class="text-slate-800">Task Title</label>
    <div class="text-xs text-slate-500">
      <span>Due: Oct 26</span>
      <PriorityBadge priority="high" />
    </div>
  </div>
  <div class="flex items-center gap-2">
    <Button variant="ghost" size="icon">...</Button> {/* Edit */}
    <Button variant="ghost" size="icon">...</Button> {/* Delete */}
  </div>
</div>

// Completed State
<div class="... opacity-60">
  <label class="... line-through text-slate-500">Task Title</label>
  ...
</div>
```

### Add/Edit Task Form (React Hook Form + Zod + shadcn)
```tsx
// Standard pattern within a Modal/Dialog
<FormField
  control={form.control}
  name="title"
  render={({ field }) => (
    <FormItem>
      <FormLabel>Task</FormLabel>
      <FormControl>
        <Input placeholder="e.g., Finish the design guide" {...field} />
      </FormControl>
      <FormMessage />
    </FormItem>
  )}
/>
```

## Page-Specific Layouts

### Main List View (`/`)
```tsx
// Centered layout, max-width
<div className="max-w-3xl mx-auto">
  {/* Header */}
  <div className="flex justify-between items-center py-8">
    <h1 className="text-2xl font-medium">My Tasks</h1>
    <Button>Add Task</Button>
  </div>

  {/* Filters */}
  <div className="flex gap-2 border-b pb-4 mb-4">
    {/* Filter/Sort controls */}
  </div>

  {/* Task List */}
  <div className="space-y-2">
    {/* Overdue Section (optional) */}
    {overdueTasks.map(task => <TodoItemRow key={task.id} />)}
    {/* Today's Tasks Section */}
    {pendingTasks.map(task => <TodoItemRow key={task.id} />)}
  </div>

  {/* Empty State */}
  <div className="text-center py-16">
    <CheckCircle className="mx-auto h-12 w-12 text-slate-300" />
    <h3 className="mt-4 text-lg font-medium">All clear!</h3>
    <p className="mt-1 text-slate-500">You have no pending tasks.</p>
  </div>
</div>
```

### Task Edit Modal
A `Dialog` component that contains the `AddTaskForm`.
```tsx
// Two-column layout for details
<div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
  <FormField name="priority" />
  <FormField name="dueDate" />
</div>
<FormField name="notes" as={Textarea} />

<DialogFooter>
  <Button variant="outline">Cancel</Button>
  <Button type="submit">Save Task</Button>
</DialogFooter>
```

### Completed View (`/completed`)
```tsx
// Similar to main view, but grouped by date
<div className="max-w-3xl mx-auto">
  <h1 className="text-2xl font-medium py-8">Completed Tasks</h1>
  
  {/* Date Group */}
  <div>
    <h2 className="text-sm font-medium text-slate-500 my-4">Today</h2>
    <div className="space-y-2">
      {completedToday.map(task => <TodoItemRow key={task.id} />)}
    </div>
  </div>

  {/* Date Group */}
  <div>
    <h2 className="text-sm font-medium text-slate-500 my-4">Yesterday</h2>
    ...
  </div>
</div>
```

## Interactive States

### Hover States
-   **Task Rows**: `hover:bg-stone-50 transition-colors`
-   **Buttons**: Built into shadcn variants. `ghost` buttons should be more prominent on hover.
-   **Links**: `hover:text-emerald-600 transition-colors`

### Focus States
-   **All Interactive Elements**: `focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2`
-   **Form Inputs**: `focus:border-emerald-500`

### Loading States
-   **Task List**: `<Skeleton />` components matching the `TodoItemRow` layout.
-   **Buttons**: Spinner inside button + disabled state + "Saving..." text.

### Error States
-   **Form Fields**: Red border (`border-destructive`) + error message below.
-   **API Errors**: Toast notification with a failure message.

## Accessibility Requirements

### WCAG AA Compliance
-   **Color Contrast**: ≥4.5:1 for normal text, ≥3:1 for large text. This must be checked for priority badges.
-   **Focus Indicators**: Visible focus rings on all interactive elements.
-   **Keyboard Navigation**: Full functionality via keyboard. Tabbing moves logically through tasks and their actions.
-   **Screen Readers**:
    -   Checkbox state changes ("checked", "not checked") must be announced.
    -   Completed tasks should be announced as "completed".
    -   Use `aria-live` regions for toast notifications.

### Form Accessibility
-   **Labels**: Every input (`<Input>`, `<Checkbox>`, `<Select>`) must have an associated `<Label>`.
-   **Error Messages**: Use `aria-invalid` and link errors with `aria-describedby`.
-   **Button Labels**: Use explicit text or `aria-label` for icon-only buttons (e.g., `aria-label="Delete task"`).

## Iconography Guidelines

Instead of product images, we rely heavily on icons for actions and status. Use the `lucide-react` library.

-   **Add**: `Plus`
-   **Edit**: `Pencil`
-   **Delete**: `Trash2`
-   **Complete**: `Check` or `CheckCircle`
-   **Priority**: `Flag` or different sized circles.
-   **Due Date**: `Calendar`
-   **Empty State**: `ClipboardCheck` or `PartyPopper`

## Component Reuse Patterns

### Required shadcn/ui Components
-   `Button` (all variants)
-   `Input`, `Label`, `Textarea`
-   `Checkbox`
-   `Form`, `FormField`, `FormItem`, `FormLabel`, `FormControl`, 'FormMessage'
-   `Dialog` (for Add/Edit modal)
-   `Select` or `RadioGroup` (for priority)
-   `Skeleton`
-   `Toast`, `useToast`

### Custom Components (Build These)
-   `TodoItemRow` - Reusable task display.
-   `PriorityBadge` - Displays priority with correct color.
- an`AddTaskForm` - The form for creating/editing tasks.
-   `TaskList` - The container for `TodoItemRow` items, handles empty/loading states.
-   `TaskListSkeleton` - Loading state for the task list.

## Definition of Done (UI Checklist)

Before marking any UI feature complete, verify:

### Layout & Spacing
-   [ ] 8pt grid spacing used throughout.
-   [ ] Consistent padding on pages and modals.
-   [ ] No custom margins/padding outside design tokens.

### Typography & Colors
-   [ ] Only approved color tokens used.
-   [ ] Task states (completed, overdue) are visually distinct and use correct tokens.
-   [ ] WCAG AA contrast ratios met on all text and badges.

### Interactive States
-   [ ] Hover states on all interactive elements (rows, buttons, links).
-   [ ] Focus rings are visible and styled correctly.
-   [ ] Loading states implemented for list fetching and form submission.
-   [ ] Empty states are designed and functional.
-   [ ] Error states (form fields, API toasts) are styled and functional.

### Component Usage
-   [ ] Only approved shadcn/ui components used.
-   [ ] Consistent button variants and sizes for actions.
-   [ ] Forms properly integrated with React Hook Form + Zod.

### Responsive Design
-   [ ] Mobile-first styling looks clean and usable.
-   [ ] Layout adjusts correctly at `sm` and `lg` breakpoints.
-   [ ] Touch targets are ≥44px on mobile.

### Accessibility
-   [ ] Keyboard navigation is logical and complete.
-   [ ] Screen reader correctly announces task status and actions.
-   [ ] Form labels and error messages are properly associated.
-   [ ] All icons have `aria-label` or are `aria-hidden`.