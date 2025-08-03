import type { Priority } from "@/types/task";
import { cn } from "@/lib/utils";

interface PriorityBadgeProps {
  priority: Priority;
  className?: string;
}

export function PriorityBadge({ priority, className }: PriorityBadgeProps) {
  if (priority === 'none') {
    return null;
  }

  const variants = {
    high: "bg-red-100 text-red-800",
    medium: "bg-yellow-100 text-yellow-800", 
    low: "bg-blue-100 text-blue-800",
  };

  const labels = {
    high: "High",
    medium: "Medium",
    low: "Low",
  };

  return (
    <span 
      className={cn(
        "inline-flex items-center px-2 py-1 rounded-full text-xs font-medium",
        variants[priority],
        className
      )}
    >
      {labels[priority]}
    </span>
  );
}