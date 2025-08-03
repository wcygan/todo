"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { taskClient } from "@/lib/client";
import { 
  CreateTaskRequestSchema, 
  GetAllTasksRequestSchema, 
  UpdateTaskRequestSchema,
  DeleteTaskRequestSchema,
  type Task as BackendTask
} from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";
import { create } from "@bufbuild/protobuf";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { TaskList } from "@/components/task-list";
import { AddTaskForm } from "@/components/add-task-form";
import { backendToFrontend } from "@/types/task";
import type { Task } from "@/types/task";
import type { TaskFormData } from "@/lib/validation/task";
import { toast } from "sonner";
import { Plus } from "lucide-react";

export default function Home() {
  const [isAddTaskOpen, setIsAddTaskOpen] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);
  const queryClient = useQueryClient();

  // Query for getting all tasks
  const { 
    data: tasksResponse, 
    isLoading, 
    error,
    refetch 
  } = useQuery({
    queryKey: ["tasks"],
    queryFn: async () => {
      const request = create(GetAllTasksRequestSchema, {});
      return await taskClient.getAllTasks(request);
    },
  });

  // Mutation for creating tasks
  const createTaskMutation = useMutation({
    mutationFn: async (description: string) => {
      const request = create(CreateTaskRequestSchema, { description });
      return await taskClient.createTask(request);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
      toast.success("Task created successfully");
    },
    onError: (error) => {
      toast.error(`Failed to create task: ${error.message}`);
    },
  });

  // Mutation for deleting tasks
  const deleteTaskMutation = useMutation({
    mutationFn: async (id: string) => {
      const request = create(DeleteTaskRequestSchema, { id });
      return await taskClient.deleteTask(request);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
      toast.success("Task deleted successfully");
    },
    onError: (error) => {
      toast.error(`Failed to delete task: ${error.message}`);
    },
  });

  // Mutation for updating task completion
  const updateTaskMutation = useMutation({
    mutationFn: async ({ id, description, completed }: { id: string; description: string; completed: boolean }) => {
      const request = create(UpdateTaskRequestSchema, { id, description, completed });
      return await taskClient.updateTask(request);
    },
    onMutate: async ({ id, completed }) => {
      // Cancel any outgoing refetches (so they don't overwrite our optimistic update)
      await queryClient.cancelQueries({ queryKey: ["tasks"] });

      // Snapshot the previous value
      const previousTasks = queryClient.getQueryData(["tasks"]);

      // Optimistically update to the new value
      queryClient.setQueryData(["tasks"], (old: any) => {
        if (!old?.tasks) return old;
        
        return {
          ...old,
          tasks: old.tasks.map((task: any) => 
            task.id === id 
              ? { 
                  ...task, 
                  completed,
                  updatedAt: completed 
                    ? { seconds: BigInt(Math.floor(Date.now() / 1000)), nanos: 0 }
                    : task.updatedAt
                }
              : task
          ),
        };
      });

      // Return a context object with the snapshotted value
      return { previousTasks };
    },
    onError: (err, variables, context) => {
      // If the mutation fails, use the context returned from onMutate to roll back
      if (context?.previousTasks) {
        queryClient.setQueryData(["tasks"], context.previousTasks);
      }
      toast.error(`Failed to update task: ${err.message}`);
    },
    onSuccess: () => {
      // Refetch to ensure we have the latest data from the server
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
    },
  });

  const handleToggleComplete = (taskId: string) => {
    const task = tasks.find(t => t.id === taskId);
    if (!task) {
      toast.error("Task not found");
      return;
    }
    
    // Toggle completion status
    updateTaskMutation.mutate({
      id: task.id,
      description: task.title,
      completed: !task.isCompleted
    });
  };

  const handleCreateTask = async (data: TaskFormData) => {
    await createTaskMutation.mutateAsync(data.title);
    setIsAddTaskOpen(false);
  };

  const handleEditTask = (task: Task) => {
    setEditingTask(task);
  };

  const handleUpdateTask = async (data: TaskFormData) => {
    // TODO: Implement when backend supports task updates
    toast.info("Task editing coming soon!");
    setEditingTask(null);
  };

  const handleDeleteTask = (id: string) => {
    deleteTaskMutation.mutate(id);
  };

  // Convert backend tasks to frontend format
  const tasks: Task[] = tasksResponse?.tasks.map(backendToFrontend) || [];

  if (error) {
    return (
      <div className="min-h-screen bg-white">
        <div className="max-w-3xl mx-auto">
          <div className="py-8">
            <div className="bg-red-100 border border-slate-200 text-red-800 p-6 rounded-lg shadow-sm">
              <p className="text-slate-900">Error loading tasks: {error.message}</p>
              <Button 
                onClick={() => refetch()}
                variant="destructive"
                className="mt-4"
              >
                Retry
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-3xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center py-8">
          <h1 className="text-2xl font-medium text-slate-900">My Tasks</h1>
          
          <Dialog open={isAddTaskOpen} onOpenChange={setIsAddTaskOpen}>
            <DialogTrigger asChild>
              <Button className="bg-emerald-500 hover:bg-emerald-600">
                <Plus className="h-4 w-4 mr-2" />
                Add Task
              </Button>
            </DialogTrigger>
            <DialogContent className="p-6 sm:max-w-[425px]">
              <DialogHeader>
                <DialogTitle>Add New Task</DialogTitle>
              </DialogHeader>
              <AddTaskForm 
                onSubmit={handleCreateTask}
                onCancel={() => setIsAddTaskOpen(false)}
                isSubmitting={createTaskMutation.isPending}
              />
            </DialogContent>
          </Dialog>
        </div>

        {/* Filters */}
        <div className="flex gap-2 border-b border-slate-200 pb-4 mb-4">
          <Button 
            variant="ghost" 
            size="sm"
            className="text-slate-600 hover:text-slate-900"
          >
            All
          </Button>
          <Button 
            variant="ghost" 
            size="sm"
            className="text-slate-600 hover:text-slate-900"
          >
            Active
          </Button>
          <Button 
            variant="ghost" 
            size="sm"
            className="text-slate-600 hover:text-slate-900"
          >
            Completed
          </Button>
        </div>

        {/* Task List */}
        <div className="space-y-2">
          <TaskList
            tasks={tasks}
            isLoading={isLoading}
            onToggleComplete={handleToggleComplete}
            onEdit={handleEditTask}
            onDelete={handleDeleteTask}
          />
        </div>

        {/* Edit Task Dialog */}
        {editingTask && (
          <Dialog open={!!editingTask} onOpenChange={() => setEditingTask(null)}>
            <DialogContent className="p-6 sm:max-w-[425px]">
              <DialogHeader>
                <DialogTitle>Edit Task</DialogTitle>
              </DialogHeader>
              <AddTaskForm 
                onSubmit={handleUpdateTask}
                onCancel={() => setEditingTask(null)}
                initialData={editingTask}
                isSubmitting={false}
              />
            </DialogContent>
          </Dialog>
        )}

        {/* Test Page Link */}
        <div className="py-8 text-center">
          <a
            href="/test"
            className="text-slate-600 hover:text-emerald-600 transition-colors"
          >
            → Go to API Test Page
          </a>
        </div>
      </div>
    </div>
  );
}
