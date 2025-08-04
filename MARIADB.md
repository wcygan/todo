# MariaDB Integration Guide

This document describes the MariaDB deployment and integration with the Todo application using the MariaDB Operator.

## Overview

The Todo application uses MariaDB as its primary database, deployed via the [MariaDB Operator](https://github.com/mariadb-operator/mariadb-operator) for cloud-native database management. The operator provides automated provisioning, backup, high availability, and lifecycle management.

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌──────────────────┐
│   Frontend      │    │    Backend       │    │    MariaDB       │
│   Next.js       │◄──►│    Go/ConnectRPC │◄──►│    Operator      │
│   Port: 3000    │    │    Port: 8080    │    │    Port: 3306    │
└─────────────────┘    └──────────────────┘    └──────────────────┘
                                               │
                                               ▼
                                        ┌──────────────────┐
                                        │  MariaDB Pod     │
                                        │  todo-mariadb-0  │
                                        │  1Gi Storage     │
                                        └──────────────────┘
```

## Deployment Components

### MariaDB Operator (Helm Chart)
- **Version**: 0.38.1
- **Namespace**: `mariadb-system`
- **Components**:
  - Main operator controller
  - Certificate controller
  - Webhook server
- **Features**: CRD management, automated backup, monitoring integration

### MariaDB Instance
- **Image**: MariaDB 11.4.2
- **Storage**: 1Gi persistent volume
- **Namespace**: `todo-app`
- **Resources**:
  - **Requests**: 100m CPU, 128Mi memory
  - **Limits**: 300m CPU, 512Mi memory

### Database Configuration
- **Database Name**: `todoapp`
- **Character Set**: `utf8mb4`
- **Collation**: `utf8mb4_unicode_ci`
- **User**: `todoapp`
- **Max Connections**: 100

## Connection Details

### Service Endpoints
```yaml
# Internal cluster access
Host: todo-mariadb.todo-app.svc.cluster.local
Port: 3306

# Headless service for StatefulSet
Host: todo-mariadb-internal.todo-app.svc.cluster.local
Port: 3306
```

### Credentials
```yaml
# Database user credentials (stored in Kubernetes secrets)
Username: todoapp
Password: todouser123  # k8s secret: todo-user-password
Database: todoapp

# Root credentials (for administrative tasks)
Root Password: rootpass123  # k8s secret: todo-mariadb-root
```

## Kubernetes Resources

### Custom Resources (CRDs)
```bash
# MariaDB instance
kubectl get mariadb todo-mariadb -n todo-app

# Database
kubectl get database todo-db -n todo-app

# User
kubectl get user todo-user -n todo-app

# Permissions
kubectl get grant todo-grant -n todo-app
```

### Standard Resources
```bash
# Pods
kubectl get pods -n todo-app -l app.kubernetes.io/name=mariadb

# Services
kubectl get services -n todo-app | grep mariadb

# Storage
kubectl get pvc -n todo-app | grep mariadb
```

## Database Schema

The MariaDB instance is initialized with optimal settings for the Todo application:

### Configuration Highlights
```ini
[mariadb]
bind-address=*
default_storage_engine=InnoDB
binlog_format=row
innodb_autoinc_lock_mode=2
max_allowed_packet=256M

# Performance settings
innodb_buffer_pool_size=128M
innodb_log_file_size=64M

# Character set
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci

# Connection settings
max_connections=100
```

### User Permissions
The `todoapp` user has the following grants on the `todoapp` database:
- SELECT, INSERT, UPDATE, DELETE
- CREATE, DROP, ALTER, INDEX
- No GRANT OPTION (security best practice)

## Operations

### Connecting to MariaDB

#### From within the cluster:
```bash
kubectl exec -it todo-mariadb-0 -n todo-app -- \
  mariadb -u todoapp -ptodouser123 todoapp
```

#### Port forwarding for external access:
```bash
kubectl port-forward -n todo-app svc/todo-mariadb 3306:3306
mysql -h 127.0.0.1 -P 3306 -u todoapp -ptodouser123 todoapp
```

### Monitoring

#### Check MariaDB instance status:
```bash
kubectl get mariadb -n todo-app
```

**Expected Output:**
```
NAME           READY   STATUS    PRIMARY POD      AGE
todo-mariadb   True    Running   todo-mariadb-0   10m
```

#### Check database resources:
```bash
kubectl get database,user,grant -n todo-app
```

#### Monitor pod health:
```bash
kubectl describe pod todo-mariadb-0 -n todo-app
kubectl logs todo-mariadb-0 -n todo-app
```

### Backup and Recovery

The MariaDB Operator supports automated backups:

```yaml
# Enable backup in mariadb-instance.yaml
backup:
  enabled: true
  schedule: "0 2 * * *"  # Daily at 2 AM
  storage:
    size: 10Gi
```

#### Manual backup:
```bash
kubectl exec todo-mariadb-0 -n todo-app -- \
  mysqldump -u root -prootpass123 todoapp > backup.sql
```

### Scaling and High Availability

#### Scale to multiple replicas:
```yaml
spec:
  replicas: 3  # Enable primary-replica setup
  galera:
    enabled: true  # Enable Galera clustering
```

#### Check replica status:
```bash
kubectl get mariadb todo-mariadb -n todo-app -o yaml | grep -A 10 status
```

## Troubleshooting

### Common Issues

#### 1. Pod Not Starting
```bash
# Check pod events
kubectl describe pod todo-mariadb-0 -n todo-app

# Check operator logs
kubectl logs -n mariadb-system deployment/mariadb-operator
```

#### 2. Connection Issues
```bash
# Test connectivity from a test pod
kubectl run mysql-client --image=mysql:8.0 -it --rm --restart=Never -- \
  mysql -h todo-mariadb.todo-app.svc.cluster.local -u todoapp -ptodouser123
```

#### 3. Storage Issues
```bash
# Check PVC status
kubectl get pvc -n todo-app

# Check storage class
kubectl get storageclass
```

#### 4. Operator Issues
```bash
# Check operator status
kubectl get pods -n mariadb-system

# Check CRDs
kubectl get crd | grep mariadb
```

### Performance Tuning

#### Resource Adjustment
```yaml
resources:
  requests:
    cpu: 200m      # Increase for higher load
    memory: 256Mi  # Increase for larger datasets
  limits:
    cpu: 500m
    memory: 1Gi
```

#### Storage Optimization
```yaml
storage:
  size: 10Gi              # Increase as needed
  storageClassName: fast-ssd  # Use SSD for better performance
```

## Integration with Backend

### Go Database Connection
```go
// Database connection string for the backend
dsn := "todoapp:todouser123@tcp(todo-mariadb.todo-app.svc.cluster.local:3306)/todoapp?charset=utf8mb4&parseTime=True&loc=Local"

db, err := sql.Open("mysql", dsn)
if err != nil {
    log.Fatal("Failed to connect to database:", err)
}
```

### Environment Variables
```yaml
# Backend deployment environment
env:
- name: DATABASE_HOST
  value: "todo-mariadb.todo-app.svc.cluster.local"
- name: DATABASE_PORT
  value: "3306"
- name: DATABASE_NAME
  value: "todoapp"
- name: DATABASE_USER
  valueFrom:
    secretKeyRef:
      name: todo-user-password
      key: username
- name: DATABASE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: todo-user-password
      key: password
```

## Security Considerations

### Secrets Management
- Database passwords stored in Kubernetes secrets
- No plaintext credentials in manifests
- Root access limited to administrative operations
- User permissions follow principle of least privilege

### Network Security
- ClusterIP services (not exposed externally)
- Internal DNS resolution only
- Can be enhanced with NetworkPolicies for traffic control

### Encryption
- Data encrypted at rest (depends on storage class)
- TLS connections can be enabled for data in transit
- Secret data encrypted in etcd

## Monitoring and Metrics

### Health Checks
The MariaDB operator provides built-in health monitoring:
- Liveness probes ensure container health
- Readiness probes ensure service availability
- Custom health checks via operator

### Metrics Collection
```yaml
# Enable metrics (requires Prometheus operator)
metrics:
  enabled: true
  serviceMonitor:
    enabled: true
```

### Performance Monitoring
```sql
-- Check slow queries
SELECT * FROM information_schema.PROCESSLIST;

-- Check database size
SELECT 
    table_schema AS 'Database',
    ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) AS 'Size (MB)'
FROM information_schema.tables 
GROUP BY table_schema;
```

## Future Enhancements

### Planned Improvements
1. **Automated Backups**: Configure scheduled backups to S3/GCS
2. **High Availability**: Enable Galera clustering for multi-master setup
3. **Connection Pooling**: Integrate with ProxySQL for connection management
4. **Monitoring**: Full Prometheus/Grafana integration
5. **Security**: TLS encryption and enhanced authentication

### Migration Path
When migrating from in-memory storage to MariaDB:
1. Update backend configuration
2. Add database migration scripts
3. Test data persistence
4. Implement connection pooling
5. Add database health checks to backend

This MariaDB setup provides a solid foundation for the Todo application with room for growth and production-ready features.