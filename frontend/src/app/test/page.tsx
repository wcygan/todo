"use client";

import { useState } from "react";
import { taskClient } from "@/lib/client";
import { 
  CreateTaskRequestSchema, 
  GetAllTasksRequestSchema, 
  DeleteTaskRequestSchema 
} from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";
import { create } from "@bufbuild/protobuf";
import { Button } from "@/components/ui/button";
import { ArrowLeft, Play, Loader2 } from "lucide-react";
import Link from "next/link";

export default function TestPage() {
  const [response, setResponse] = useState<string>("");
  const [loading, setLoading] = useState(false);
  const [taskId, setTaskId] = useState<string>("");

  // Helper function to safely stringify protobuf objects
  const safeStringify = (obj: unknown) => {
    return JSON.stringify(obj, (key, value) => {
      if (typeof value === 'bigint') {
        return value.toString();
      }
      return value;
    }, 2);
  };

  const testCreateTask = async () => {
    setLoading(true);
    try {
      const request = create(CreateTaskRequestSchema, {
        description: "Test task from frontend",
      });
      
      const result = await taskClient.createTask(request);
      setResponse(safeStringify(result));
      
      // Store task ID for deletion test
      if (result.task?.id) {
        setTaskId(result.task.id);
      }
    } catch (error) {
      setResponse(`Error: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const testGetAllTasks = async () => {
    setLoading(true);
    try {
      const request = create(GetAllTasksRequestSchema, {});
      const result = await taskClient.getAllTasks(request);
      setResponse(safeStringify(result));
    } catch (error) {
      setResponse(`Error: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  const testDeleteTask = async () => {
    if (!taskId) {
      setResponse("No task ID available. Create a task first.");
      return;
    }
    
    setLoading(true);
    try {
      const request = create(DeleteTaskRequestSchema, {
        id: taskId,
      });
      
      const result = await taskClient.deleteTask(request);
      setResponse(safeStringify(result));
      setTaskId(""); // Clear the stored ID
    } catch (error) {
      setResponse(`Error: ${error}`);
    } finally {
      setLoading(false);
    }
  };

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
          <h1 className="text-2xl font-medium text-slate-900">API Test Page</h1>
        </div>

        {/* Test Controls */}
        <div className="space-y-4 mb-6">
          <Button
            onClick={testCreateTask}
            disabled={loading}
            className="bg-emerald-500 hover:bg-emerald-600 text-white"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Creating...
              </>
            ) : (
              <>
                <Play className="h-4 w-4 mr-2" />
                Create Task
              </>
            )}
          </Button>

          <Button
            onClick={testGetAllTasks}
            disabled={loading}
            variant="outline"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Loading...
              </>
            ) : (
              "Get All Tasks"
            )}
          </Button>

          <Button
            onClick={testDeleteTask}
            disabled={loading || !taskId}
            variant="destructive"
          >
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Deleting...
              </>
            ) : (
              `Delete Task ${taskId ? `(${taskId.substring(0, 8)}...)` : "(No ID)"}`
            )}
          </Button>
        </div>

        {/* Response Display */}
        {response && (
          <div className="border border-slate-200 rounded-lg p-6 bg-stone-50">
            <h2 className="text-lg font-medium text-slate-900 mb-4">Response:</h2>
            <pre className="bg-white border border-slate-200 p-4 rounded text-sm overflow-auto text-slate-800 shadow-sm">
              {response}
            </pre>
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