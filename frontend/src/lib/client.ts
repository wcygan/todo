import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import { TaskService } from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";

// Always use the API route - Next.js will handle the proxy
const transport = createConnectTransport({
  baseUrl: '/api',
});

// Create the client
export const taskClient = createClient(TaskService, transport);