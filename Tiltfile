# Tilt configuration following best practices for Go + Next.js development
allow_k8s_contexts(['minikube', 'docker-desktop', 'kind-kind', 'orbstack'])

# Load extensions for enhanced development
load('ext://restart_process', 'docker_build_with_restart')

# Create namespace
k8s_yaml(['k8s/development/namespace.yaml'])

# Apply services and configurations
k8s_yaml([
    'k8s/rbac/rbac.yaml',
    'k8s/development/secrets.yaml',
    'k8s/development/backend-configmap.yaml',
    'k8s/development/frontend-configmap.yaml',
    'k8s/development/mysql-operator.yaml',
    'k8s/development/mysql-init.yaml',
    'k8s/development/backend-deployment.yaml',
    'k8s/development/backend-service.yaml',
    'k8s/development/frontend-deployment.yaml',
    'k8s/development/frontend-service.yaml',
])

# Backend build with live updates - Tilt best practices
local_resource(
    'backend-compile',
    'cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o ../k8s/backend/server ./cmd/server',
    deps=['backend/cmd', 'backend/internal', 'backend/go.mod', 'backend/go.sum'],
    labels=['backend']
)

docker_build_with_restart(
    'todo-backend',
    context='./k8s/backend',
    dockerfile='./k8s/backend/dockerfile.dev',
    entrypoint='/app/server',
    only=['./server'],
    live_update=[
        sync('./k8s/backend/server', '/app/server'),
    ],
)

# Frontend build with optimized caching
docker_build(
    'todo-frontend',
    context='./frontend',
    dockerfile='./k8s/frontend/dockerfile.dev',
    # Live updates for development
    live_update=[
        sync('./frontend/src', '/app/src'),
        sync('./frontend/app', '/app/app'),
        sync('./frontend/public', '/app/public'),
        sync('./frontend/package.json', '/app/package.json'),
    ],
)

# Configure resources with port forwarding
k8s_resource('backend', 
    port_forwards=['8080:8080'],
    resource_deps=['backend-compile'],
    labels=['backend']
)

k8s_resource('frontend', 
    port_forwards=['3000:3000'],
    labels=['frontend']
)

# MySQL initialization job
k8s_resource('mysql-init',
    labels=['database']
)

# MySQL cluster resources - track the InnoDBCluster custom resource
k8s_resource(
    new_name='mysql-cluster',
    objects=['mysql-cluster:InnoDBCluster:todo-app'],
    labels=['database']
)

print("✅ Tilt configured with best practices - Go live_update + Next.js hot reload + MySQL!")