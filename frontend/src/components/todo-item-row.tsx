import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import { PriorityBadge } from "@/components/priority-badge";
import type { Task } from "@/types/task";
import { Pencil, Trash2 } from "lucide-react";
import { cn } from "@/lib/utils";

interface TodoItemRowProps {
  task: Task;
  onToggleComplete: (taskId: string) => void;
  onEdit: (task: Task) => void;
  onDelete: (taskId: string) => void;
}

export function TodoItemRow({ task, onToggleComplete, onEdit, onDelete }: TodoItemRowProps) {
  const formattedDueDate = task.dueDate ? 
    new Date(task.dueDate).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }) : 
    null;

  const isOverdue = task.dueDate && !task.isCompleted && new Date(task.dueDate) < new Date();

  return (
    <div className={cn(
      "flex items-center p-3 hover:bg-stone-50 rounded-lg transition-colors",
      task.isCompleted && "opacity-60"
    )}>
      {/* Checkbox */}
      <Checkbox
        id={`task-${task.id}`}
        checked={task.isCompleted}
        onCheckedChange={() => onToggleComplete(task.id)}
        aria-label={task.isCompleted ? "Mark as incomplete" : "Mark as complete"}
      />

      {/* Title & Details */}
      <div className="flex-1 mx-4">
        <label 
          htmlFor={`task-${task.id}`}
          className={cn(
            "block text-slate-800 cursor-pointer",
            task.isCompleted && "line-through text-slate-500"
          )}
        >
          {task.title}
        </label>
        
        {(formattedDueDate || task.priority !== 'none') && (
          <div className="flex items-center gap-2 mt-1">
            {formattedDueDate && (
              <span className={cn(
                "text-xs",
                isOverdue ? "text-red-600" : "text-slate-500"
              )}>
                Due: {formattedDueDate}
              </span>
            )}
            {task.priority !== 'none' && <PriorityBadge priority={task.priority} />}
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2">
        <Button 
          variant="ghost" 
          size="icon"
          onClick={() => onEdit(task)}
          aria-label="Edit task"
        >
          <Pencil className="h-4 w-4" />
        </Button>
        
        <Button 
          variant="ghost" 
          size="icon"
          onClick={() => onDelete(task.id)}
          aria-label="Delete task"
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}