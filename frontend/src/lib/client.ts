import { createConnectTransport } from "@connectrpc/connect-web";
import { createClient } from "@connectrpc/connect";
import { TaskService } from "@buf/wcygan_todo.bufbuild_es/task/v1/task_pb.js";

// Create the transport for HTTP/1.1 or HTTP/2
const transport = createConnectTransport({
  baseUrl: "http://backend-service:8080",
});

// Create the client
export const taskClient = createClient(TaskService, transport);