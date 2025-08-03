# Frontend in Tilt Hangs on Load: Debugging Summary

The application is currently unavailable when running via `tilt up`. The frontend page at `http://localhost:3000` hangs indefinitely upon loading.

## Root Cause

The core issue is that the **backend pods are unable to establish a connection with the MySQL database service** within the Kubernetes cluster. This prevents the backend from becoming fully healthy, which in turn causes the frontend to hang while it waits for a response that never comes.

## Investigation Steps & Findings

1.  **Initial State**: All pods (frontend, backend, database) were confirmed to be running. However, new backend pods were failing their readiness checks.

2.  **Frontend Misconfiguration**: Identified and corrected the frontend's API client, which was incorrectly configured to connect to `http://localhost:8080` instead of the correct in-cluster Kubernetes service name, `http://backend-service:8080`.

3.  **Missing Network Policy**: Discovered that no `NetworkPolicy` was defined for the `todo-app` namespace. This effectively firewalled all services from each other. A policy was created and applied to explicitly allow:
    *   `frontend` → `backend` (on port 8080)
    *   `backend` → `mysql-cluster-instances` (on port 3306)

4.  **Persistent Connection Failure**: Despite the fixes, the problem persists. We attempted to `exec` into a running backend pod to manually test the database connection using `mysql-client`.

5.  **Final Blocker**: The database connection test from within the pod failed with an unexpected `exit code 139`. This indicates a crash or segmentation fault within the container when the `mysql` command is executed, pointing to a deeper, unresolved issue within the backend container's environment or its interaction with the database service.

## Current Status

The issue is blocked pending further investigation into the `exit code 139` error occurring inside the backend container.
