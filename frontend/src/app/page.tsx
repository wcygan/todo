"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTransport } from "@connectrpc/connect-query";
import { 
  CreateTaskRequestSchema, 
  GetAllTasksRequestSchema, 
  DeleteTaskRequestSchema,
  type Task
} from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";
import { 
  createTask as createTaskQuery,
  getAllTasks as getAllTasksQuery,
  deleteTask as deleteTaskQuery
} from "@buf/wcygan_todo.connectrpc_query-es/task/v1/task-TaskService_connectquery.js";
import { create } from "@bufbuild/protobuf";

export default function Home() {
  const [newTaskDescription, setNewTaskDescription] = useState("");
  const transport = useTransport();
  const queryClient = useQueryClient();

  // Query for getting all tasks
  const { 
    data: tasksResponse, 
    isLoading, 
    error,
    refetch 
  } = useQuery({
    queryKey: ["tasks"],
    queryFn: () => {
      const request = create(GetAllTasksRequestSchema, {});
      return getAllTasksQuery(request, { transport });
    },
  });

  // Mutation for creating tasks
  const createTaskMutation = useMutation({
    mutationFn: (description: string) => {
      const request = create(CreateTaskRequestSchema, { description });
      return createTaskQuery(request, { transport });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
      setNewTaskDescription("");
    },
  });

  // Mutation for deleting tasks
  const deleteTaskMutation = useMutation({
    mutationFn: (id: string) => {
      const request = create(DeleteTaskRequestSchema, { id });
      return deleteTaskQuery(request, { transport });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["tasks"] });
    },
  });

  const handleCreateTask = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTaskDescription.trim()) return;
    createTaskMutation.mutate(newTaskDescription);
  };

  const handleDeleteTask = (id: string) => {
    deleteTaskMutation.mutate(id);
  };

  const tasks = tasksResponse?.tasks || [];

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-2xl mx-auto px-4">
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
            <p>Error loading tasks: {error.message}</p>
            <button 
              onClick={() => refetch()}
              className="mt-2 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 py-8">
      <div className="max-w-2xl mx-auto px-4">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Todo App with TanStack Query</h1>
        
        {/* Add Task Form */}
        <form onSubmit={handleCreateTask} className="mb-8">
          <div className="flex gap-2">
            <input
              type="text"
              value={newTaskDescription}
              onChange={(e) => setNewTaskDescription(e.target.value)}
              placeholder="Enter a new task..."
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              disabled={createTaskMutation.isPending}
            />
            <button
              type="submit"
              disabled={createTaskMutation.isPending || !newTaskDescription.trim()}
              className="px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
            >
              {createTaskMutation.isPending ? "Adding..." : "Add Task"}
            </button>
          </div>
          {createTaskMutation.error && (
            <p className="text-red-500 text-sm mt-2">
              Error: {createTaskMutation.error.message}
            </p>
          )}
        </form>

        {/* Loading State */}
        {isLoading && (
          <div className="text-center py-8">
            <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
            <p className="mt-2 text-gray-500">Loading tasks...</p>
          </div>
        )}

        {/* Tasks List */}
        {!isLoading && (
          <div className="space-y-2">
            {tasks.length === 0 ? (
              <p className="text-gray-500 text-center py-8">No tasks yet. Add one above!</p>
            ) : (
              tasks.map((task: Task) => (
                <div
                  key={task.id}
                  className="flex items-center justify-between p-4 bg-white rounded-lg shadow-sm border"
                >
                  <div className="flex-1">
                    <p className="text-gray-900">{task.description}</p>
                    <p className="text-sm text-gray-500">
                      Created: {task.createdAt ? new Date(task.createdAt.toDate()).toLocaleString() : "Unknown"}
                    </p>
                  </div>
                  <button
                    onClick={() => handleDeleteTask(task.id)}
                    disabled={deleteTaskMutation.isPending}
                    className="ml-4 px-3 py-1 text-sm bg-red-500 text-white rounded hover:bg-red-600 disabled:opacity-50"
                  >
                    {deleteTaskMutation.isPending ? "Deleting..." : "Delete"}
                  </button>
                </div>
              ))
            )}
          </div>
        )}

        {/* Delete Error */}
        {deleteTaskMutation.error && (
          <div className="mt-4 bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
            <p>Error deleting task: {deleteTaskMutation.error.message}</p>
          </div>
        )}

        {/* Test Page Link */}
        <div className="mt-8 text-center">
          <a
            href="/test"
            className="text-blue-500 hover:text-blue-600 underline"
          >
            → Go to API Test Page
          </a>
        </div>
      </div>
    </div>
  );
}
