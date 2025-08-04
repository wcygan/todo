# Skaffold Deployment Guide

This document describes how to use Skaffold for developing and deploying the Todo application to Kubernetes.

## Overview

Skaffold provides a streamlined development workflow for Kubernetes applications with:
- Automatic image building and deployment
- Hot reloading for rapid development
- Port forwarding for local access
- Multi-environment configuration profiles

## Architecture

Our Skaffold setup deploys a two-tier application:

```
┌─────────────────┐    ┌──────────────────┐
│   Frontend      │    │    Backend       │
│   Next.js       │◄──►│    Go/ConnectRPC │
│   Port: 3000    │    │    Port: 8080    │
│   (Bun/Node)    │    │    (Distroless)  │
└─────────────────┘    └──────────────────┘
```

**Service Communication:**
- Frontend → Backend: `http://backend-service:8080`
- Local Access: Port forwarding to `localhost:3000` and `localhost:8080`

## Usage

### Development Workflow

#### 1. Start Development Mode (Recommended)
```bash
skaffold dev --profile=dev
```

This command:
- Builds Docker images locally
- Deploys to Kubernetes
- Sets up port forwarding
- Watches for file changes and redeploys automatically
- Streams logs from all services

**Expected Output:**
```
Generating tags...
 - todo-backend -> todo-backend:c2efbd3-dirty
 - todo-frontend -> todo-frontend:c2efbd3
Starting build...
Build [todo-backend] succeeded
Build [todo-frontend] succeeded
Starting deploy...
 - namespace/todo-app created
 - deployment.apps/backend created
 - service/backend-service created
 - deployment.apps/frontend created
 - service/frontend-service created
Deployments stabilized in 16.108 seconds
Port forwarding service/frontend-service in namespace todo-app, remote port 3000 -> 127.0.0.1:3000
Port forwarding service/backend-service in namespace todo-app, remote port 8080 -> 127.0.0.1:8080
```

#### 2. One-time Deployment
```bash
skaffold run --profile=dev
```

Builds and deploys once without watching for changes.

#### 3. Clean Up Resources
```bash
skaffold delete --profile=dev
```

Removes all deployed Kubernetes resources.

### Production Deployment

```bash
skaffold run --profile=production
```

Uses cluster-based image building and deployment optimized for production environments.

## Configuration Profiles

### Development Profile (`dev`)
- **Image Building**: Local Docker daemon
- **Image Push**: Disabled (`push: false`)
- **Port Forwarding**: Automatic for both services
- **Context**: `orbstack` (local Kubernetes)
- **Hot Reload**: Enabled with file synchronization

### Production Profile (`production`)
- **Image Building**: Cluster-based
- **Image Push**: Enabled
- **Port Forwarding**: Disabled
- **Context**: Production cluster

## Verification

### 1. Check Deployment Status

```bash
kubectl get pods -n todo-app
```

**Expected Output:**
```
NAME                       READY   STATUS    RESTARTS   AGE
backend-5c8c9b488d-tk52d   1/1     Running   0          2m
frontend-c4b9b4df8-bw9k2   1/1     Running   0          2m
```

### 2. Verify Services

```bash
kubectl get services -n todo-app
```

**Expected Output:**
```
NAME               TYPE        CLUSTER-IP        EXTERNAL-IP   PORT(S)    AGE
backend-service    ClusterIP   192.168.194.246   <none>        8080/TCP   2m
frontend-service   ClusterIP   192.168.194.201   <none>        3000/TCP   2m
```

### 3. Test Health Endpoints

**Backend Health Check:**
```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{"status":"healthy","service":"todo-backend"}
```

**Frontend Access:**
```bash
curl -I http://localhost:3000
```

**Expected Response:**
```
HTTP/1.1 200 OK
Content-Type: text/html; charset=utf-8
```

### 4. Application Functionality Test

1. **Navigate to:** http://localhost:3000/test
2. **Test API Integration:** Click "Create Task" button
3. **Verify:** Tasks appear in the list, confirming frontend-backend communication

### 5. Monitor Application Logs

```bash
# All services
skaffold dev --tail

# Specific service
kubectl logs -n todo-app deployment/backend -f
kubectl logs -n todo-app deployment/frontend -f
```

## Extendability

### Adding New Services

1. **Add Docker Build Configuration:**
```yaml
build:
  artifacts:
  - image: todo-new-service
    context: ./new-service
    docker:
      dockerfile: Dockerfile
```

2. **Create Kubernetes Manifests:**
```bash
mkdir -p k8s/new-service
# Add deployment.yaml and service.yaml
```

3. **Update Manifest List:**
```yaml
manifests:
  rawYaml:
  - k8s/new-service/deployment.yaml
  - k8s/new-service/service.yaml
```

### Environment-Specific Configuration

#### Custom Profile Example
```yaml
profiles:
- name: staging
  build:
    cluster: {}
  deploy:
    kubectl:
      flags:
        global: ["--context=staging-cluster"]
  manifests:
    rawYaml:
    - k8s/staging/
```

#### Usage:
```bash
skaffold run --profile=staging
```

### File Synchronization for Development

Current sync configuration for hot reloading:

```yaml
build:
  artifacts:
  - image: todo-backend
    sync:
      manual:
      - src: "backend/**/*.go"
        dest: .
  - image: todo-frontend
    sync:
      manual:
      - src: "**/*.js"
        dest: .
      - src: "**/*.jsx"
        dest: .
      - src: "**/*.ts"
        dest: .
      - src: "**/*.tsx"
        dest: .
```

**Add More File Types:**
```yaml
sync:
  manual:
  - src: "**/*.css"
    dest: .
  - src: "**/*.json"
    dest: .
```

## Troubleshooting

### Common Issues

#### 1. Image Pull Errors
**Problem:** `ErrImagePull` or `ImagePullBackOff`

**Solution:**
- Ensure Docker context is set correctly: `docker context use orbstack`
- Verify `imagePullPolicy: Never` in deployment manifests for local development
- Check image exists: `docker images | grep todo`

#### 2. Port Forwarding Conflicts
**Problem:** Port already in use

**Solution:**
```bash
# Kill existing port forwards
pkill -f "kubectl port-forward"

# Use different ports
skaffold dev --port-forward --port-forward-ports=3001:3000,8081:8080
```

#### 3. Health Check Failures
**Problem:** Pods fail readiness/liveness probes

**Solution:**
```bash
# Check pod logs
kubectl logs -n todo-app deployment/backend

# Test health endpoint directly
kubectl exec -n todo-app deployment/backend -- /app/server &
sleep 2
kubectl exec -n todo-app deployment/backend -- pkill server
```

#### 4. Build Context Issues
**Problem:** Build fails with "context" errors

**Solution:**
- Verify Dockerfile paths in `skaffold.yaml`
- Check `.dockerignore` isn't excluding necessary files
- Use `--verbosity=debug` for detailed build logs

### Debug Commands

```bash
# Detailed deployment info
skaffold diagnose

# Verbose logging
skaffold dev --verbosity=debug

# Skip image building (use existing)
skaffold dev --cache-artifacts=false

# Force rebuild
skaffold dev --no-prune=false --cache-artifacts=false
```

## Performance Optimization

### Build Optimization
```yaml
build:
  local:
    useBuildkit: true
    concurrency: 2
```

### Resource Management
Current resource configuration per service:

**Backend:**
```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "50m"
  limits:
    memory: "128Mi"
    cpu: "200m"
```

**Frontend:**
```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "500m"
```

### Image Optimization
- Backend uses distroless base image (44.8MB)
- Frontend uses multi-stage builds with Node.js Alpine (771MB)
- Both images leverage Docker layer caching

## Integration with Development Tools

### IDE Integration
- Configure your IDE to trigger `skaffold dev` on file saves
- Use `skaffold debug` for remote debugging capabilities

### CI/CD Integration
```yaml
# GitHub Actions example
- name: Deploy with Skaffold
  run: skaffold run --profile=production --default-repo=${{ secrets.REGISTRY }}
```

### Local Development Stack
```bash
# Complete local development setup
docker context use orbstack
skaffold dev --profile=dev --port-forward
```

This provides a complete local development environment with:
- ✅ Hot reloading for code changes
- ✅ Automatic port forwarding
- ✅ Health check monitoring  
- ✅ Log streaming
- ✅ Service mesh networking

## Security Considerations

All deployments include:
- **Non-root users**: Both containers run as non-root
- **Read-only filesystems**: Where possible
- **Dropped capabilities**: All unnecessary capabilities removed
- **Resource limits**: Prevent resource exhaustion
- **Health checks**: Ensure service availability
- **Network policies**: Can be added for traffic control

This Skaffold configuration provides a production-ready foundation while maintaining an excellent developer experience.