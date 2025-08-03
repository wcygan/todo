"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { taskClient } from "@/lib/client";
import { 
  CreateTaskRequestSchema, 
  GetAllTasksRequestSchema, 
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

  // TODO: Add task completion toggle mutation when backend supports it
  const handleToggleComplete = (taskId: string) => {
    // For now, just show a toast - will implement when backend supports completion toggle
    toast.info("Task completion toggle coming soon!");
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
      <div className="min-h-screen bg-white py-8">
        <div className="max-w-3xl mx-auto px-4">
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
            <p>Error loading tasks: {error.message}</p>
            <Button 
              onClick={() => refetch()}
              variant="destructive"
              className="mt-2"
            >
              Retry
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white">
      <div className="max-w-3xl mx-auto px-4">
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
            <DialogContent>
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

        {/* Task List */}
        <TaskList
          tasks={tasks}
          isLoading={isLoading}
          onToggleComplete={handleToggleComplete}
          onEdit={handleEditTask}
          onDelete={handleDeleteTask}
        />

        {/* Edit Task Dialog */}
        {editingTask && (
          <Dialog open={!!editingTask} onOpenChange={() => setEditingTask(null)}>
            <DialogContent>
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
        <div className="mt-8 pb-8 text-center">
          <a
            href="/test"
            className="text-emerald-600 hover:text-emerald-700 transition-colors underline"
          >
            → Go to API Test Page
          </a>
        </div>
      </div>
    </div>
  );
}
