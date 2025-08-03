# Local Testing Guide for Next.js + Kubernetes

## Quick Start

```bash
# 1. Start Tilt (this handles everything)
tilt up

# 2. Access the application
open http://localhost:3000
```

That's it! Tilt automatically:
- Builds Docker images
- Deploys to Kubernetes
- Sets up port forwarding
- Enables hot reload

## Architecture Overview

```
Browser → localhost:3000 → Next.js → /api/* → Backend Service → MySQL
         (port forward)   (Frontend)  (proxy)   (internal)
```

## Manual Testing (Without Tilt)

If you prefer manual control:

### 1. Build Images
```bash
# Build frontend
docker build -t todo-frontend:latest -f k8s/frontend/Dockerfile frontend/

# Build backend  
docker build -t todo-backend:latest -f k8s/backend/Dockerfile backend/
```

### 2. Deploy to Kubernetes
```bash
# Apply all manifests
kubectl apply -k k8s/development/
```

### 3. Wait for Pods
```bash
# Check pod status
kubectl get pods -n todo-app -w

# Wait for all pods to be ready
kubectl wait --for=condition=ready pod --all -n todo-app --timeout=300s
```

### 4. Set Up Port Forwarding
```bash
# In separate terminals (or use & to background)

# Frontend (BFF)
kubectl port-forward svc/frontend-service -n todo-app 3000:3000

# Backend (optional - for direct API testing)
kubectl port-forward svc/backend-service -n todo-app 8080:8080

# Database (optional - for debugging)
kubectl port-forward svc/mysql-service -n todo-app 3306:3306
```

## Testing the BFF Pattern

### 1. Browser Test
```bash
# Open the app
open http://localhost:3000

# Should see the todo list UI
# Creating/deleting tasks should work
```

### 2. API Route Test
```bash
# Test the BFF proxy
curl http://localhost:3000/api/task.v1.TaskService/GetAllTasks \
  -H "Content-Type: application/json" \
  -d '{}'

# Should return tasks from backend through Next.js
```

### 3. Direct Backend Test (Optional)
```bash
# Only if you port-forwarded 8080
curl http://localhost:8080/task.v1.TaskService/GetAllTasks \
  -H "Content-Type: application/json" \
  -d '{}'
```

## Verify BFF is Working

### Check Network Flow
```bash
# 1. Watch frontend logs
kubectl logs -f deployment/frontend -n todo-app

# 2. In another terminal, watch backend logs  
kubectl logs -f deployment/backend -n todo-app

# 3. Create a task in the browser
# You should see:
# - Request hit frontend first
# - Frontend proxies to backend
# - Backend processes and returns
```

### Inspect API Route
```bash
# The frontend should have this route
kubectl exec deployment/frontend -n todo-app -- ls -la /app/src/app/api/
# Should show: [...path]/route.ts
```

## Common Issues

### Frontend Can't Reach Backend
```bash
# Check if backend is running
kubectl get pods -n todo-app | grep backend

# Check service DNS
kubectl exec deployment/frontend -n todo-app -- nslookup backend-service

# Test connection from frontend pod
kubectl exec deployment/frontend -n todo-app -- curl http://backend-service:8080/health
```

### Port Forwarding Issues
```bash
# Kill existing port forwards
pkill -f "kubectl port-forward"

# Check what's using ports
lsof -i :3000
lsof -i :8080
```

## Development Workflow

### With Tilt (Recommended)
1. `tilt up` - Start everything
2. Edit code - Changes auto-sync
3. See results immediately
4. `tilt down` - Clean up

### Without Tilt
1. Edit code
2. Rebuild image: `docker build ...`
3. Delete pod: `kubectl delete pod ...`
4. Wait for new pod
5. Test changes

## Production vs Development

### Development (Current)
- Port forwarding for access
- Hot reload enabled
- Debug logging
- Single replicas

### Production
- Ingress controller for access
- Optimized builds
- Multiple replicas
- Horizontal pod autoscaling

## Verify Everything is Working

Run the health check:
```bash
# If using Tilt
tilt trigger health-check

# Or manually
curl -sf http://localhost:3000/api/health && echo "✅ Frontend OK" || echo "❌ Frontend Failed"
curl -sf http://localhost:8080/health && echo "✅ Backend OK" || echo "❌ Backend Failed"
```

## Next Steps

1. **Test the app**: Create, complete, and delete tasks
2. **Check logs**: Verify requests flow through BFF
3. **Modify code**: Test hot reload functionality
4. **Run tests**: `tilt trigger backend-test`