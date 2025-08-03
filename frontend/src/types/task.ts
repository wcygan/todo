import type { Task as BackendTask } from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";

// Priority levels as defined in design.md
export type Priority = 'low' | 'medium' | 'high' | 'none';

// Frontend task type with all design.md fields
export interface Task {
  id: string;
  title: string;
  isCompleted: boolean;
  priority: Priority;
  dueDate?: string; // ISO date string
  notes?: string;
  createdAt: string; // ISO date string
  completedAt?: string; // ISO date string
}

// Adapter functions to convert between backend and frontend schemas
export function backendToFrontend(backendTask: BackendTask): Task {
  return {
    id: backendTask.id,
    title: backendTask.description, // Backend 'description' maps to frontend 'title'
    isCompleted: backendTask.completed,
    priority: 'none', // Default priority (backend doesn't have this yet)
    dueDate: undefined, // Backend doesn't have this yet
    notes: undefined, // Backend doesn't have this yet
    createdAt: backendTask.createdAt ? 
      new Date(Number(backendTask.createdAt.seconds) * 1000).toISOString() : 
      new Date().toISOString(),
    completedAt: backendTask.completed && backendTask.updatedAt ? 
      new Date(Number(backendTask.updatedAt.seconds) * 1000).toISOString() : 
      undefined,
  };
}

// For future when backend supports all fields
export function frontendToBackend(task: Task): Partial<BackendTask> {
  return {
    id: task.id,
    description: task.title,
    completed: task.isCompleted,
    // Priority, dueDate, notes will be added when backend supports them
  };
}