# Tilt Development Environment Guide

This document provides comprehensive guidance for using the Tilt development environment for the todo application.

## Overview

The Tiltfile provides a complete development environment with:
- **Hot reload** for Go backend (Air) and Next.js frontend (Turbopack)
- **Protocol buffer automation** with dependency management
- **Database integration** with MySQL operator
- **Testing integration** with manual triggers
- **Debug tools** and performance monitoring
- **Local registry** for faster builds

## Quick Start

### Prerequisites Setup
```bash
# Run the automated setup script
./tilt-setup.sh

# Or manually install:
# - Docker Desktop
# - Minikube
# - kubectl, helm, buf CLI, tilt
# - Go 1.24.2+
# - Bun package manager
```

### Basic Commands
```bash
# Start all services
tilt up

# Start specific services only
tilt up --services=backend,frontend

# Enable monitoring stack
tilt up --monitoring=true

# Stop all services
tilt down

# View logs
tilt logs <resource-name>
```

## Service Architecture

### Core Services
- **Backend**: Go ConnectRPC service on port 8080
- **Frontend**: Next.js application on port 3000  
- **Database**: MySQL cluster on port 3306
- **Ingress**: nginx-ingress for routing

### Resource Dependencies
```
protobuf-generate
    ↓
mysql-operator → mysql-cluster → backend → frontend → ingress
```

## Development Features

### Protocol Buffer Integration
```bash
# Auto-generates code when .proto files change
tilt trigger protobuf-generate

# Updates dependencies manually  
tilt trigger protobuf-deps-update
```

### Live Updates
- **Backend**: Air hot reload - no container restarts
- **Frontend**: Turbopack HMR - instant updates
- **Optimized syncing**: Only changed files are synced

### Database Management
```bash
# Initialize database schema
tilt trigger mysql-init-schema

# Open database shell
tilt trigger db-shell

# Create backup
tilt trigger db-backup

# View database logs
tilt trigger logs-database
```

## Testing Integration

### Available Test Commands
```bash
# Backend unit tests
tilt trigger backend-test-unit

# Backend integration tests  
tilt trigger backend-test-integration

# Frontend unit tests
tilt trigger frontend-test-unit

# Frontend linting and type checking
tilt trigger frontend-lint

# Run all tests
tilt trigger test-all
```

### Test Configuration
- **Backend**: Go race detection, coverage reports
- **Frontend**: Bun test runner with type checking
- **Integration**: Full HTTP API testing
- **Manual triggers**: Tests don't auto-run to avoid noise

## Debug Tools

### Logging
```bash
# Stream service logs
tilt trigger logs-backend
tilt trigger logs-frontend  
tilt trigger logs-database

# Or use tilt logs directly
tilt logs backend
```

### Health Monitoring
```bash
# Check all service health
tilt trigger health-check

# Individual service checks via curl
curl http://localhost:8080/health  # Backend
curl http://localhost:3000         # Frontend
```

### Performance Tools
```bash
# Clean build caches
tilt trigger clean-cache

# Reset entire environment
tilt trigger reset-environment
```

## Advanced Configuration

### Selective Service Deployment
```bash
# Backend only
tilt up --services=backend,mysql-cluster

# Frontend development (assumes backend running elsewhere)
tilt up --services=frontend

# Database only
tilt up --services=mysql-operator,mysql-cluster
```

### Environment Variables
The Tiltfile reads from:
- `./k8s/config/backend-configmap.yaml`
- `./k8s/config/frontend-configmap.yaml`
- `./k8s/config/secrets.yaml`

### Local Registry
- Automatically starts `localhost:5000` registry
- Reduces image pull times
- Enables offline development

## Performance Optimizations

### Build Caching
- **Docker layer caching**: Optimized multi-stage builds
- **Go module caching**: Persistent mod cache
- **Bun install caching**: Frozen lockfile installs

### Live Updates
- **Incremental syncing**: Only changed files
- **Smart rebuilds**: Triggered by specific file patterns
- **Hot reload**: Minimal container restarts

### Resource Management
- **Parallel updates**: Max 3 concurrent updates
- **Resource quotas**: Optimized for local development
- **Dependency ordering**: Efficient startup sequence

## Monitoring Stack

Enable with `--monitoring=true`:
```bash
tilt up --monitoring=true

# Access monitoring tools
tilt trigger port-forward-grafana    # http://localhost:3001
tilt trigger port-forward-prometheus # http://localhost:9090
```

## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check resource status
tilt get all

# View specific logs
tilt logs <resource-name>

# Restart specific resource
tilt trigger restart-<resource-name>
```

#### Database Connection Issues
```bash
# Verify MySQL is ready
kubectl get pods -n todo-app
kubectl logs mysql-cluster-0 -n todo-app

# Reset database
tilt trigger reset-environment
```

#### Build Failures
```bash
# Clean caches
tilt trigger clean-cache

# Check Docker daemon
docker system info

# Verify local registry
docker ps --filter name=registry
```

### Performance Issues

#### Slow Builds
```bash
# Enable local registry
docker run -d -p 5000:5000 --name registry registry:2

# Clean Docker system
docker system prune -af --volumes
```

#### Memory Issues
```bash
# Check resource usage
docker stats
kubectl top pods -n todo-app

# Reduce resource limits in k8s/development/kustomization.yaml
```

## File Structure

```
/
├── Tiltfile              # Main Tilt configuration
├── tilt-setup.sh         # Automated setup script
├── TILT.md              # This documentation
├── backup/              # Database backup storage
├── k8s/                 # Kubernetes manifests
│   ├── backend/
│   │   ├── dockerfile       # Production Dockerfile
│   │   └── dockerfile.dev   # Development Dockerfile with Air
│   ├── frontend/
│   │   ├── dockerfile       # Production Dockerfile  
│   │   └── dockerfile.dev   # Development Dockerfile with Bun
│   └── development/     # Kustomize development config
├── backend/
│   └── .air.toml        # Air configuration for hot reload
├── frontend/            # Next.js application
└── proto/              # Protocol buffer definitions
```

## Best Practices

### Development Workflow
1. Start with `tilt up` for full environment
2. Use `tilt trigger test-all` before committing
3. Monitor logs with `tilt logs <service>`
4. Clean caches periodically
5. Use health checks to verify setup

### Performance Tips
- Use selective deployment for focused development
- Enable local registry for faster builds
- Monitor resource usage with `docker stats`
- Clean caches when experiencing issues

### Debugging Strategy
1. Check service health first
2. Review logs for specific errors
3. Verify database connectivity
4. Test individual components
5. Reset environment as last resort

## Resources

- [Tilt Documentation](https://docs.tilt.dev/)
- [Air (Go Hot Reload)](https://github.com/air-verse/air)
- [Turbopack (Next.js)](https://turbo.build/pack)
- [MySQL Operator](https://mysql.github.io/mysql-operator/)
- [Protocol Buffers](https://protobuf.dev/)

## Support

For issues with this Tilt configuration:
1. Check the troubleshooting section above
2. Review Tilt logs: `tilt logs`
3. Verify prerequisites with `./tilt-setup.sh`
4. Reset environment: `tilt trigger reset-environment`