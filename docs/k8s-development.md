# Kubernetes Development with Tilt

This document provides a comprehensive guide for deploying and developing the todo application on Kubernetes using Tilt as the development toolkit. The setup includes a full-stack todo application with Go ConnectRPC backend, Next.js frontend, and MySQL database managed by the MySQL Operator.

## Architecture Overview

### Technology Stack
- **Backend**: Go 1.24.2 + ConnectRPC + Protocol Buffers
- **Frontend**: Next.js 15 + React 19 + TypeScript + Tailwind CSS 4
- **Database**: MySQL 9.0 with MySQL Operator for Kubernetes
- **Development**: Tilt + Minikube/Docker Desktop
- **Schema Management**: Buf Schema Registry (buf.build/wcygan/todo)

### Service Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   Database      │
│   Next.js:3000  │───▶│   Go:8080       │───▶│   MySQL:3306    │
│                 │    │   ConnectRPC    │    │   InnoDBCluster │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Core Components
- **Frontend Service**: Next.js application with ConnectRPC Web client
- **Backend Service**: Go ConnectRPC service with HTTP/2 support
- **Database Service**: MySQL InnoDBCluster with automatic failover
- **Protocol Buffers**: Type-safe communication via buf.build registry

## Prerequisites

### Required Tools
```bash
# Kubernetes cluster (choose one)
minikube start --memory=4096 --cpus=2
# OR
docker desktop (enable Kubernetes)

# Development tools
brew install tilt-dev/tap/tilt
brew install kubernetes-cli
brew install helm

# For protocol buffers
brew install bufbuild/buf/buf
```

### MySQL Operator Installation
```bash
# Install MySQL Operator
helm repo add mysql-operator https://mysql.github.io/mysql-operator/
helm repo update
helm install mysql-operator mysql-operator/mysql-operator \
  --namespace mysql-operator --create-namespace

# Verify installation
kubectl get pods -n mysql-operator
```

### Ingress Controller (Optional)
```bash
# For external access via ingress
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml

# Wait for ingress controller
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

## Quick Start

### 1. Environment Setup
```bash
# Clone and navigate to project
cd /path/to/todo

# Run setup script
./tilt-setup.sh

# Configure local DNS (for ingress)
echo "$(minikube ip) todo.local api.todo.local" | sudo tee -a /etc/hosts
```

### 2. Start Development Environment
```bash
# Start Tilt development environment
tilt up

# Access Tilt UI
open http://localhost:10350
```

### 3. Access Services
```bash
# Port forwarding (automatic via Tilt)
# Frontend: http://localhost:3000
# Backend: http://localhost:8080
# Database: mysql://root:dev-password@localhost:3306

# Via ingress (if configured)
# Frontend: http://todo.local
# Backend: http://api.todo.local
```

## Development Workflow

### Protocol Buffer Development
```bash
# Modify .proto files in proto/
vim proto/task/v1/task.proto

# Tilt automatically triggers:
# 1. buf generate
# 2. Go package updates
# 3. TypeScript package updates
# 4. Service restarts

# Manual protobuf operations
tilt trigger protobuf-gen
tilt trigger protobuf-deps
```

### Backend Development
```bash
# Make changes to Go code
vim backend/internal/handler/task.go

# Tilt provides:
# - Live code sync to container
# - Air hot reload (no rebuild)
# - Automatic service restart
# - Health check monitoring

# Manual operations
tilt trigger backend-test
tilt trigger backend-build
```

### Frontend Development
```bash
# Make changes to React/Next.js code
vim frontend/src/components/task-list.tsx

# Tilt provides:
# - Hot module replacement
# - Turbopack fast builds
# - Live dependency updates
# - Browser refresh

# Manual operations
tilt trigger frontend-test
tilt trigger frontend-build
```

### Database Development
```bash
# Schema changes (via backend migrations)
vim backend/migrations/001_initial.sql

# Database operations
tilt trigger db-migrate
tilt trigger db-backup
tilt trigger db-restore

# Direct access
kubectl port-forward service/todo-mysql 3306:3306
mysql -h localhost -u root -p todoapp
```

## Tilt Configuration

### Resource Organization
The Tilt configuration organizes resources into logical groups:

- **`data`**: MySQL database and initialization
- **`api`**: Backend Go service
- **`web`**: Frontend Next.js application
- **`tools`**: Development utilities and testing

### Live Update Features
```python
# Backend live updates
docker_build('todo-backend', './backend',
    dockerfile='./k8s/backend/Dockerfile.dev',
    live_update=[
        sync('./backend', '/app'),
        run('go build -o /app/tmp/server ./cmd/server')
    ]
)

# Frontend live updates
docker_build('todo-frontend', './frontend',
    dockerfile='./k8s/frontend/Dockerfile.dev',
    live_update=[
        sync('./frontend/src', '/app/src'),
        sync('./frontend/public', '/app/public')
    ]
)
```

### Performance Optimizations
- **Parallel builds**: Up to 3 concurrent resource updates
- **Local registry**: Faster image pulls with localhost:5000
- **Build caching**: Docker layer optimization
- **Selective sync**: Only changed files trigger updates

## Database Management

### MySQL Operator Configuration
```yaml
# Development configuration
apiVersion: mysql.oracle.com/v2
kind: InnoDBCluster
metadata:
  name: todo-mysql
spec:
  instances: 1                    # Single instance for dev
  router:
    instances: 1
  datadirVolumeClaimTemplate:
    resources:
      requests:
        storage: 10Gi
  resources:
    requests:
      memory: "512Mi"
      cpu: "250m"
    limits:
      memory: "1Gi"
      cpu: "500m"
```

### Schema Management
```sql
-- Database initialization (automatic via k8s job)
CREATE DATABASE IF NOT EXISTS todoapp;
USE todoapp;

CREATE TABLE tasks (
  id VARCHAR(36) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  is_completed BOOLEAN DEFAULT FALSE,
  priority ENUM('low', 'medium', 'high', 'none') DEFAULT 'none',
  due_date DATETIME NULL,
  notes TEXT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  completed_at DATETIME NULL
);
```

### Backup and Restore
```bash
# Create backup
tilt trigger db-backup

# List backups
ls backup/

# Restore from backup
tilt trigger db-restore BACKUP_FILE=backup-20240803.sql
```

## Testing Integration

### Automated Testing
```bash
# Run all tests
tilt trigger test-all

# Backend tests
tilt trigger backend-test

# Frontend tests
tilt trigger frontend-test

# Integration tests
tilt trigger integration-test
```

### Manual Testing
```bash
# Health checks
tilt trigger health-check

# Load testing
tilt trigger load-test

# API testing
curl http://localhost:8080/task.v1.TaskService/GetAllTasks
```

## Debugging and Monitoring

### Logging
```bash
# View logs via Tilt UI
open http://localhost:10350

# Direct kubectl access
kubectl logs -f deployment/todo-backend
kubectl logs -f deployment/todo-frontend
kubectl logs -f deployment/todo-mysql
```

### Debug Configuration
```bash
# Enable debug mode
tilt trigger debug-mode

# Backend debugging (port 2345)
dlv attach --listen=:2345 --headless --api-version=2

# Frontend debugging (port 9229)
node --inspect-brk=0.0.0.0:9229
```

### Resource Monitoring
```bash
# Resource usage
kubectl top pods
kubectl describe pod todo-backend-xxx

# Tilt resource status
tilt get resources
```

## Production Considerations

### Scaling Configuration
```yaml
# Production backend deployment
spec:
  replicas: 3
  resources:
    requests:
      memory: "1Gi"
      cpu: "500m"
    limits:
      memory: "2Gi"
      cpu: "1000m"
```

### High Availability Database
```yaml
# Production MySQL cluster
spec:
  instances: 3                    # Multi-instance cluster
  router:
    instances: 2                  # Load balancing
  backupSchedules:
    - name: daily-backup
      schedule: "0 2 * * *"       # Daily at 2 AM
      enabled: true
```

### Security Enhancements
- Network policies for pod-to-pod communication
- RBAC with minimal required permissions
- TLS certificates for production ingress
- Secret management via Kubernetes secrets
- Non-root container execution

### CI/CD Integration
```yaml
# GitHub Actions example
name: Deploy to Kubernetes
on:
  push:
    branches: [main]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy with Tilt
        run: |
          tilt ci
          kubectl rollout status deployment/todo-backend
```

## Troubleshooting

### Common Issues

#### 1. MySQL Operator Not Ready
```bash
# Check operator status
kubectl get pods -n mysql-operator

# Check operator logs
kubectl logs -n mysql-operator deployment/mysql-operator

# Restart operator
kubectl rollout restart deployment/mysql-operator -n mysql-operator
```

#### 2. Service Connection Issues
```bash
# Check service endpoints
kubectl get endpoints

# Test internal connectivity
kubectl run debug --image=busybox -it --rm -- nslookup todo-backend

# Check network policies
kubectl get networkpolicies
```

#### 3. Build Failures
```bash
# Clear Tilt cache
tilt down --delete-namespaces
tilt up

# Check Docker daemon
docker ps
docker system df

# Restart local registry
docker restart local-registry
```

#### 4. Resource Limits
```bash
# Check resource usage
kubectl top nodes
kubectl top pods

# Increase Minikube resources
minikube delete
minikube start --memory=8192 --cpus=4
```

### Performance Tuning

#### Optimize Build Times
```python
# Tiltfile optimizations
update_settings(max_parallel_updates=3)
default_registry('localhost:5000')

# Use build caching
docker_build('app', '.', 
    cache_from=['app:latest'],
    build_args={'BUILDKIT_INLINE_CACHE': '1'}
)
```

#### Optimize Resource Usage
```yaml
# Pod resource tuning
resources:
  requests:
    memory: "256Mi"    # Start small
    cpu: "100m"
  limits:
    memory: "512Mi"    # Allow headroom
    cpu: "200m"
```

## Future Enhancements

### Planned Improvements
1. **Observability**: Prometheus metrics + Grafana dashboards
2. **Tracing**: OpenTelemetry distributed tracing
3. **Service Mesh**: Istio for advanced traffic management
4. **GitOps**: ArgoCD for declarative deployments
5. **Secrets Management**: External Secrets Operator
6. **Cost Optimization**: Vertical Pod Autoscaler

### Migration Path
1. **Development → Staging**: Multi-environment Tilt configurations
2. **Staging → Production**: Helm charts with environment overlays
3. **Single Cluster → Multi-Cluster**: Federation and cross-cluster deployments

## Resources

### Documentation
- [Tilt Documentation](https://docs.tilt.dev/)
- [MySQL Operator Documentation](https://github.com/mysql/mysql-operator)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Buf Documentation](https://docs.buf.build/)

### Configuration Files
- `Tiltfile` - Main Tilt configuration
- `k8s/` - Kubernetes manifests
- `TILT.md` - Detailed Tilt usage guide
- `tilt-setup.sh` - Environment setup script

### Support
- Project issues: GitHub Issues
- Tilt community: [Tilt Slack](https://tilt.dev/community)
- Kubernetes community: [Kubernetes Slack](https://slack.k8s.io/)

---

This setup provides a production-ready development environment that scales from local development to production deployment while maintaining consistency across environments.