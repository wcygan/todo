"use client";

import { useQuery } from "@tanstack/react-query";
import { taskClient } from "@/lib/client";
import { GetAllTasksRequestSchema } from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";
import { create } from "@bufbuild/protobuf";
import { Button } from "@/components/ui/button";
import { TodoItemRow } from "@/components/todo-item-row";
import { backendToFrontend } from "@/types/task";
import type { Task } from "@/types/task";
import { ArrowLeft } from "lucide-react";
import Link from "next/link";

export default function CompletedPage() {
  // Query for getting all tasks
  const { 
    data: tasksResponse, 
    isLoading, 
    error 
  } = useQuery({
    queryKey: ["tasks"],
    queryFn: async () => {
      const request = create(GetAllTasksRequestSchema, {});
      return await taskClient.getAllTasks(request);
    },
  });

  // Convert backend tasks to frontend format and filter completed
  const allTasks: Task[] = tasksResponse?.tasks.map(backendToFrontend) || [];
  const completedTasks = allTasks.filter(task => task.isCompleted);

  // Group completed tasks by date
  const groupedTasks = completedTasks.reduce((groups, task) => {
    const completedDate = task.completedAt ? new Date(task.completedAt) : new Date();
    const dateKey = completedDate.toDateString();
    
    if (!groups[dateKey]) {
      groups[dateKey] = [];
    }
    groups[dateKey].push(task);
    return groups;
  }, {} as Record<string, Task[]>);

  // Sort date groups (most recent first)
  const sortedDateKeys = Object.keys(groupedTasks).sort((a, b) => 
    new Date(b).getTime() - new Date(a).getTime()
  );

  const formatDateHeader = (dateString: string) => {
    const date = new Date(dateString);
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(today.getDate() - 1);

    if (date.toDateString() === today.toDateString()) {
      return "Today";
    } else if (date.toDateString() === yesterday.toDateString()) {
      return "Yesterday";
    } else {
      return date.toLocaleDateString('en-US', { 
        weekday: 'long', 
        year: 'numeric', 
        month: 'long', 
        day: 'numeric' 
      });
    }
  };

  if (error) {
    return (
      <div className="min-h-screen bg-white">
        <div className="max-w-3xl mx-auto">
          <div className="py-8">
            <div className="bg-red-100 border border-slate-200 text-red-800 p-6 rounded-lg shadow-sm">
              <p className="text-slate-900">Error loading tasks: {error.message}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="min-h-screen bg-white">
        <div className="max-w-3xl mx-auto">
          <div className="py-8">
            <div className="text-center">Loading completed tasks...</div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-3xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 py-8">
          <Link href="/">
            <Button variant="ghost" size="icon">
              <ArrowLeft className="h-4 w-4" />
            </Button>
          </Link>
          <h1 className="text-2xl font-medium text-slate-900">Completed Tasks</h1>
        </div>

        {/* Completed Tasks Grouped by Date */}
        {completedTasks.length === 0 ? (
          <div className="text-center py-16">
            <p className="text-slate-500">No completed tasks yet.</p>
          </div>
        ) : (
          <div className="space-y-6">
            {sortedDateKeys.map((dateKey) => (
              <div key={dateKey}>
                <h2 className="text-sm font-medium text-slate-500 mb-4">
                  {formatDateHeader(dateKey)}
                </h2>
                <div className="space-y-2">
                  {groupedTasks[dateKey].map((task) => (
                    <TodoItemRow
                      key={task.id}
                      task={task}
                      onToggleComplete={() => {}} // Read-only in completed view
                      onEdit={() => {}} // Read-only in completed view
                      onDelete={() => {}} // Read-only in completed view
                    />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Back to Main Link */}
        <div className="py-8 text-center">
          <Link
            href="/"
            className="text-emerald-600 hover:text-emerald-600 transition-colors underline"
          >
            ← Back to Tasks
          </Link>
        </div>
      </div>
    </div>
  );
}