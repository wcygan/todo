import { createConnectTransport } from "@connectrpc/connect-web";
import { QueryClient } from "@tanstack/react-query";
import { createConnectQueryKey, TransportProvider } from "@connectrpc/connect-query";

// Create the transport for HTTP/1.1 or HTTP/2
const transport = createConnectTransport({
  baseUrl: "http://localhost:8080",
});

// Create and export the transport provider
export const transportProvider: TransportProvider = () => transport;

// Create query client with sensible defaults
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5000, // Data is fresh for 5 seconds
      gcTime: 10 * 60 * 1000, // Cache for 10 minutes
    },
  },
});

// Export the query key creator for convenience
export { createConnectQueryKey };