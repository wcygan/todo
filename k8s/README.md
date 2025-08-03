# Todo App Kubernetes Manifests

This directory contains comprehensive Kubernetes manifests for deploying the todo application in a development environment with production considerations.

## Directory Structure

```
k8s/
├── namespace/           # Namespace definition
├── config/             # ConfigMaps and Secrets
├── database/           # MySQL InnoDBCluster configuration
├── backend/            # Backend service manifests
├── frontend/           # Frontend service manifests
├── networking/         # Ingress and network policies
├── rbac/               # RBAC configuration
├── monitoring/         # Monitoring setup
└── development/        # Development-specific configurations
```

## Quick Start

### Prerequisites

1. **Kubernetes cluster** (Minikube for development)
2. **MySQL Operator for Kubernetes**
3. **Nginx Ingress Controller**
4. **Tilt** (optional, for development workflow)

### Installation Steps

1. **Install MySQL Operator:**
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/mysql/mysql-operator/trunk/deploy/deploy-crds.yaml
   kubectl apply -f https://raw.githubusercontent.com/mysql/mysql-operator/trunk/deploy/deploy-operator.yaml
   ```

2. **Install Nginx Ingress Controller:**
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml
   ```

3. **Deploy the application:**
   ```bash
   # Using Kustomize
   kubectl apply -k k8s/development/

   # Or manually apply in order:
   kubectl apply -f k8s/namespace/
   kubectl apply -f k8s/config/
   kubectl apply -f k8s/rbac/
   kubectl apply -f k8s/database/
   kubectl apply -f k8s/backend/
   kubectl apply -f k8s/frontend/
   kubectl apply -f k8s/networking/
   ```

4. **Configure local DNS:**
   ```bash
   # Add to /etc/hosts
   echo "$(minikube ip) todo.local api.todo.local" | sudo tee -a /etc/hosts
   ```

5. **Access the application:**
   - Frontend: http://todo.local
   - Backend API: http://api.todo.local

## Configuration Details

### Backend Configuration

- **Port:** 8080 (ConnectRPC over HTTP/2)
- **Health Checks:** gRPC reflection endpoints
- **Resource Limits:** 256Mi memory, 200m CPU (development)
- **Configuration:** Mounted from ConfigMap at `/etc/config/`

### Frontend Configuration

- **Port:** 3000 (Next.js)
- **Environment Variables:** Set via ConfigMap
- **Resource Limits:** 512Mi memory, 300m CPU (development)
- **API Connection:** Connects to `backend-service:8080`

### Database Configuration

- **MySQL InnoDBCluster:** Single instance for development
- **Storage:** 10Gi persistent volume
- **Credentials:** Stored in Kubernetes secrets
- **Initialization:** Automatic table creation via init job

## Development Workflow

### Using Tilt (Recommended)

1. **Install Tilt:** https://tilt.dev/
2. **Create Tiltfile in project root:**
   ```python
   # Build and deploy backend
   docker_build('todo-backend', './backend', dockerfile='./k8s/backend/dockerfile')
   k8s_yaml(kustomize('./k8s/development'))
   
   # Build and deploy frontend
   docker_build('todo-frontend', './frontend', dockerfile='./k8s/frontend/dockerfile')
   
   # Port forwards
   k8s_resource('frontend', port_forwards=3000)
   k8s_resource('backend', port_forwards=8080)
   ```
3. **Run:** `tilt up`

### Manual Development

1. **Build images:**
   ```bash
   # Backend
   docker build -t todo-backend:latest -f k8s/backend/dockerfile ./backend
   
   # Frontend
   docker build -t todo-frontend:latest -f k8s/frontend/dockerfile ./frontend
   ```

2. **Load images into Minikube:**
   ```bash
   minikube image load todo-backend:latest
   minikube image load todo-frontend:latest
   ```

3. **Apply manifests:**
   ```bash
   kubectl apply -k k8s/development/
   ```

## Production Considerations

### Security

- **RBAC:** Minimal permissions for service accounts
- **Network Policies:** Restrict pod-to-pod communication
- **Pod Security Standards:** Non-root containers, dropped capabilities
- **Secrets Management:** Use external secret management (e.g., Vault)

### Scalability

- **Backend:** Scale to 3+ replicas with HPA
- **Frontend:** Scale to 2+ replicas behind CDN
- **Database:** Scale to 3-node InnoDBCluster
- **Resource Limits:** Increase for production workloads

### Monitoring

- **Prometheus:** Metrics collection
- **Grafana:** Dashboard visualization
- **Alerting:** Configure alerts for failures
- **Logging:** Centralized log aggregation

### High Availability

- **Multi-AZ Deployment:** Spread across availability zones
- **Database Replication:** Master-slave MySQL setup
- **Load Balancing:** External load balancer with health checks
- **Backup Strategy:** Automated database backups

## Troubleshooting

### Common Issues

1. **MySQL Operator not found:**
   ```bash
   kubectl get pods -n mysql-operator
   ```

2. **Images not found:**
   ```bash
   # Check if images are loaded
   minikube image ls | grep todo
   ```

3. **DNS resolution:**
   ```bash
   kubectl exec -it <backend-pod> -- nslookup mysql-cluster-mysql-master
   ```

4. **Service connectivity:**
   ```bash
   kubectl port-forward svc/backend-service 8080:8080
   kubectl port-forward svc/frontend-service 3000:3000
   ```

### Logs

```bash
# Backend logs
kubectl logs -f deployment/backend -n todo-app

# Frontend logs
kubectl logs -f deployment/frontend -n todo-app

# MySQL logs
kubectl logs -f mysql-cluster-0 -n todo-app
```

## Environment Variables

### Backend

- `CONFIG_PATH`: Path to configuration file
- `DATABASE_URL`: MySQL connection string
- `ENVIRONMENT`: Deployment environment

### Frontend

- `NEXT_PUBLIC_API_URL`: Backend API URL
- `NEXT_PUBLIC_ENVIRONMENT`: Environment name
- `NODE_ENV`: Node.js environment
- `PORT`: Server port

## Resource Requirements

### Development (Minikube)

- **CPU:** 2 cores minimum
- **Memory:** 4GB minimum
- **Storage:** 20GB minimum

### Production

- **Backend:** 100m-500m CPU, 256Mi-1Gi memory per replica
- **Frontend:** 100m-300m CPU, 128Mi-512Mi memory per replica
- **Database:** 500m-2 CPU, 1Gi-8Gi memory
- **Storage:** Fast SSD for database, 100Gi+ for production

## Maintenance

### Updates

1. **Update images:**
   ```bash
   kubectl set image deployment/backend backend=todo-backend:v1.1.0 -n todo-app
   kubectl set image deployment/frontend frontend=todo-frontend:v1.1.0 -n todo-app
   ```

2. **Rolling restarts:**
   ```bash
   kubectl rollout restart deployment/backend -n todo-app
   kubectl rollout restart deployment/frontend -n todo-app
   ```

### Backup

```bash
# Database backup
kubectl exec mysql-cluster-0 -n todo-app -- mysqldump -u root -p todoapp > backup.sql

# Configuration backup
kubectl get configmaps,secrets -n todo-app -o yaml > config-backup.yaml
```