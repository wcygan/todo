# Frontend in Tilt Issue - Complete Analysis

## Issues Found and Resolved

### 1. MySQL Cluster Initialization (✅ FIXED)
**Problem:** MySQL Operator couldn't initialize the cluster due to authentication errors.
**Solution:** 
- Replaced complex MySQL Operator with simple MySQL StatefulSet
- Fixed credentials (changed from root to todouser)
- Updated service names and network policies

### 2. Frontend Hanging Issue (⚠️ REQUIRES REBUILD)
**Problem:** Frontend hangs when accessed because it's trying to connect to `backend-service:8080` from the browser.
**Root Cause:** 
- Frontend Docker image has hardcoded `backend-service:8080` URL
- Browsers can't resolve Kubernetes service names
- Server-side rendering in Next.js tries to fetch data during initial page load

**Code Changes Made:**
1. Updated `/frontend/src/lib/client.ts` to detect server vs client environment
2. Added proxy configuration in `/frontend/next.config.ts`
3. Updated configmap to use localhost for development

**Status:** Changes are made but require Docker image rebuild.

## Current State

### Working ✅
- MySQL database is running and accessible
- Backend API is fully functional (tested via curl)
- All pods are running
- Network policies are correctly configured

### Not Working ❌
- Frontend UI hangs on load (needs image rebuild with code changes)
- Tilt is not running (would automatically rebuild on code changes)

## Solution to Complete the Fix

### Option 1: Use Tilt (Recommended)
```bash
# From project root
tilt up

# Tilt will:
# - Detect the code changes
# - Rebuild the frontend image
# - Deploy it automatically
# - Provide port forwarding
```

### Option 2: Manual Docker Build
```bash
# Build frontend with updated code
cd frontend
docker build -t todo-frontend:latest .

# Load into Kubernetes (if using kind/minikube)
kind load docker-image todo-frontend:latest

# Restart deployment
kubectl rollout restart deployment/frontend -n todo-app
```

### Option 3: Run Frontend Locally
```bash
# From frontend directory
cd frontend
export BACKEND_URL=http://localhost:8080
bun install
bun dev

# In another terminal, port-forward the backend
kubectl port-forward svc/backend-service -n todo-app 8080:8080
```

## Technical Details

### The API URL Problem
1. **Server-side:** Next.js needs `http://backend-service:8080` (Kubernetes service)
2. **Client-side:** Browser needs `http://localhost:8080` (or public URL)

### Solution Implemented
```typescript
// frontend/src/lib/client.ts
const getBaseUrl = () => {
  if (typeof window === 'undefined') {
    // Server-side: use Kubernetes service
    return process.env.BACKEND_URL || 'http://backend-service:8080';
  }
  // Client-side: use relative URL (proxied by Next.js)
  return '';
};
```

### Proxy Configuration
```typescript
// frontend/next.config.ts
{
  source: '/task.v1.TaskService/:path*',
  destination: 'http://backend-service:8080/task.v1.TaskService/:path*',
}
```

## Verification Steps

Once the frontend is rebuilt:

1. Access http://localhost:3000
2. The page should load without hanging
3. Tasks should be displayed (if any exist)
4. Creating/deleting tasks should work

## Lessons Learned

1. **Environment-specific URLs:** Always handle server vs client URL differences in Next.js
2. **MySQL Operator Complexity:** Sometimes simpler deployments work better for development
3. **Tilt Benefits:** Automatic rebuilds on code changes are essential for Kubernetes development
4. **Network Policies:** Ensure labels match between services and policies