import { TodoItemRow } from "@/components/todo-item-row";
import { Skeleton } from "@/components/ui/skeleton";
import type { Task } from "@/types/task";
import { CheckCircle } from "lucide-react";

interface TaskListProps {
  tasks: Task[];
  isLoading: boolean;
  onToggleComplete: (taskId: string) => void;
  onEdit: (task: Task) => void;
  onDelete: (taskId: string) => void;
}

function TaskListSkeleton() {
  return (
    <div className="space-y-2">
      {Array.from({ length: 3 }).map((_, index) => (
        <div key={index} className="flex items-center p-3 rounded-lg">
          <Skeleton className="h-4 w-4 rounded" />
          <div className="flex-1 mx-4">
            <Skeleton className="h-4 w-3/4 mb-2" />
            <Skeleton className="h-3 w-1/2" />
          </div>
          <div className="flex items-center gap-2">
            <Skeleton className="h-8 w-8" />
            <Skeleton className="h-8 w-8" />
          </div>
        </div>
      ))}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="text-center py-16">
      <CheckCircle className="mx-auto h-12 w-12 text-slate-300" />
      <h3 className="mt-4 text-lg font-medium text-slate-900">All clear!</h3>
      <p className="mt-1 text-slate-500">You have no pending tasks.</p>
    </div>
  );
}

export function TaskList({ tasks, isLoading, onToggleComplete, onEdit, onDelete }: TaskListProps) {
  if (isLoading) {
    return <TaskListSkeleton />;
  }

  if (tasks.length === 0) {
    return <EmptyState />;
  }

  // Group tasks by status and priority
  const activeTasks = tasks.filter(task => !task.isCompleted);
  const completedTasks = tasks.filter(task => task.isCompleted);
  const overdueTasks = activeTasks.filter(task => 
    task.dueDate && new Date(task.dueDate) < new Date()
  );
  const regularTasks = activeTasks.filter(task => 
    !task.dueDate || new Date(task.dueDate) >= new Date()
  );

  return (
    <div className="space-y-6">
      {/* Overdue Section */}
      {overdueTasks.length > 0 && (
        <div>
          <h2 className="text-sm font-medium text-red-600 mb-4">Overdue</h2>
          <div className="space-y-2">
            {overdueTasks.map((task) => (
              <TodoItemRow
                key={task.id}
                task={task}
                onToggleComplete={onToggleComplete}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        </div>
      )}

      {/* Regular Tasks Section */}
      {regularTasks.length > 0 && (
        <div>
          {overdueTasks.length > 0 && (
            <h2 className="text-sm font-medium text-slate-500 mb-4">Today</h2>
          )}
          <div className="space-y-2">
            {regularTasks.map((task) => (
              <TodoItemRow
                key={task.id}
                task={task}
                onToggleComplete={onToggleComplete}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        </div>
      )}

      {/* Empty state for active tasks when only completed tasks exist */}
      {activeTasks.length === 0 && completedTasks.length > 0 && (
        <EmptyState />
      )}

      {/* Completed Tasks Section */}
      {completedTasks.length > 0 && (
        <div>
          <h2 className="text-sm font-medium text-slate-400 mb-4">
            Completed ({completedTasks.length})
          </h2>
          <div className="space-y-2">
            {completedTasks.map((task) => (
              <TodoItemRow
                key={task.id}
                task={task}
                onToggleComplete={onToggleComplete}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  );
}