"use client";

import { useState } from "react";
import { taskClient } from "@/lib/client";
import { 
  CreateTaskRequestSchema, 
  GetAllTasksRequestSchema, 
  DeleteTaskRequestSchema 
} from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";
import { create } from "@bufbuild/protobuf";

export default function TestPage() {
  const [response, setResponse] = useState<string>("");
  const [loading, setLoading] = useState(false);
  const [taskId, setTaskId] = useState<string>("");

  const testCreateTask = async () => {
    setLoading(true);
    try {
      const request = create(CreateTaskRequestSchema, {
        description: "Test task from frontend",
      });
      
      const result = await taskClient.createTask(request);
      setResponse(JSON.stringify(result, null, 2));
      
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
      setResponse(JSON.stringify(result, null, 2));
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
      setResponse(JSON.stringify(result, null, 2));
      setTaskId(""); // Clear the stored ID
    } catch (error) {
      setResponse(`Error: ${error}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold mb-4">ConnectRPC Test</h1>
      
      <div className="space-y-4">
        <button
          onClick={testCreateTask}
          disabled={loading}
          className="bg-green-500 text-white px-4 py-2 rounded hover:bg-green-600 disabled:opacity-50"
        >
          {loading ? "Creating..." : "Create Task"}
        </button>

        <button
          onClick={testGetAllTasks}
          disabled={loading}
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600 disabled:opacity-50"
        >
          {loading ? "Loading..." : "Get All Tasks"}
        </button>

        <button
          onClick={testDeleteTask}
          disabled={loading || !taskId}
          className="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600 disabled:opacity-50"
        >
          {loading ? "Deleting..." : `Delete Task ${taskId ? `(${taskId})` : "(No ID)"}`}
        </button>
      </div>

      {response && (
        <div className="mt-6">
          <h2 className="text-lg font-semibold mb-2">Response:</h2>
          <pre className="bg-gray-100 p-4 rounded overflow-auto text-sm">
            {response}
          </pre>
        </div>
      )}
    </div>
  );
}