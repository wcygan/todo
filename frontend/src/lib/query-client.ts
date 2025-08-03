import { QueryClient } from "@tanstack/react-query";

// Create query client with sensible defaults
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5000, // Data is fresh for 5 seconds
      gcTime: 10 * 60 * 1000, // Cache for 10 minutes
    },
  },
});