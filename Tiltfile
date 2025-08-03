# Todo App Development with Tilt
# Optimized for BFF architecture with Next.js frontend and Go backend

# =============================================================================
# CONFIGURATION
# =============================================================================

allow_k8s_contexts(['minikube', 'docker-desktop', 'kind-kind', 'orbstack'])

# Check for local registry (optional but speeds up builds)
local_registry = 'localhost:5000'
if str(local('docker ps --filter "name=registry" --format "{{.Names}}}" 2>/dev/null || true', quiet=True)).strip():
    default_registry(local_registry)
    print("✅ Using local registry: " + local_registry)

# Load extensions
load('ext://restart_process', 'docker_build_with_restart')

# =============================================================================
# KUBERNETES RESOURCES
# =============================================================================

# Apply all Kubernetes manifests using kustomize
k8s_yaml(kustomize('./k8s/development'))

# =============================================================================
# BACKEND BUILD
# =============================================================================

# Compile backend locally for faster rebuilds
local_resource(
    'backend-compile',
    cmd='cd backend && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o ../k8s/backend/server ./cmd/server',
    deps=['backend/cmd', 'backend/internal', 'backend/go.mod', 'backend/go.sum'],
    labels=['backend', 'build']
)

# Backend Docker build with hot reload
docker_build_with_restart(
    'todo-backend',
    context='./k8s/backend',
    dockerfile='./k8s/backend/Dockerfile.dev',
    entrypoint='/app/server',
    only=['./server'],
    live_update=[
        sync('./k8s/backend/server', '/app/server'),
    ],
)

# =============================================================================
# FRONTEND BUILD
# =============================================================================

# Frontend Docker build with Next.js hot reload
docker_build(
    'todo-frontend',
    context='./frontend',
    dockerfile='./k8s/frontend/Dockerfile.dev',
    build_args={
        'NODE_ENV': 'development',
        'NEXT_TELEMETRY_DISABLED': '1',
    },
    live_update=[
        # Sync source files for hot reload
        sync('./frontend/src/', '/app/src/'),
        sync('./frontend/public/', '/app/public/'),
        sync('./frontend/next.config.ts', '/app/next.config.ts'),
        sync('./frontend/tailwind.config.ts', '/app/tailwind.config.ts'),
        sync('./frontend/postcss.config.mjs', '/app/postcss.config.mjs'),
        sync('./frontend/components.json', '/app/components.json'),
        
        # Handle package changes
        sync('./frontend/package.json', '/app/package.json'),
        sync('./frontend/bun.lockb', '/app/bun.lockb'),
        run('cd /app && bun install', trigger=['package.json', 'bun.lockb']),
    ],
)

# =============================================================================
# PROTOCOL BUFFERS
# =============================================================================

local_resource(
    'protobuf-generate',
    cmd='buf generate',
    deps=['./proto', './buf.gen.yaml'],
    labels=['protobuf', 'codegen'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
)

# =============================================================================
# RESOURCE CONFIGURATION
# =============================================================================

# Backend service
k8s_resource(
    'backend',
    port_forwards=['8080:8080'],
    resource_deps=['backend-compile', 'mysql'],
    labels=['backend', 'api'],
)

# Frontend service (BFF)
k8s_resource(
    'frontend',
    port_forwards=['3000:3000'],
    resource_deps=['backend'],
    labels=['frontend', 'bff'],
)

# MySQL database (StatefulSet)
k8s_resource(
    'mysql',
    port_forwards=['3306:3306'],
    labels=['database'],
)

# Configuration resources
k8s_resource(
    'mysql-init',
    objects=[
        'mysql-credentials:secret',
        'backend-secrets:secret',
        'backend-config:configmap',
        'frontend-config:configmap',
    ],
    labels=['config'],
)

# =============================================================================
# DEVELOPMENT TOOLS
# =============================================================================

# Backend tests
local_resource(
    'backend-test',
    cmd='cd backend && go test -v ./...',
    deps=['backend/'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['test', 'backend'],
)

# Frontend tests
local_resource(
    'frontend-test',
    cmd='cd frontend && bun test',
    deps=['frontend/src'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['test', 'frontend'],
)

# Frontend linting
local_resource(
    'frontend-lint',
    cmd='cd frontend && bun run lint',
    deps=['frontend/src'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['test', 'frontend'],
)

# Database shell access
local_resource(
    'mysql-shell',
    cmd='kubectl exec -it mysql-0 -n todo-app -- mysql -u todouser -p',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['database', 'debug'],
    resource_deps=['mysql'],
)

# Health check
local_resource(
    'health-check',
    cmd='''
        echo "🏥 Checking service health..."
        echo "Backend API:"
        curl -sf http://localhost:8080/health || echo "❌ Backend not ready"
        echo "\nFrontend:"
        curl -sf http://localhost:3000/api/health || echo "❌ Frontend not ready"
        echo "\nDatabase:"
        kubectl exec mysql-0 -n todo-app -- mysqladmin ping -u todouser -p 2>/dev/null || echo "❌ Database not ready"
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['debug', 'health'],
    resource_deps=['backend', 'frontend', 'mysql'],
)

# Graceful shutdown helper
local_resource(
    'graceful-shutdown',
    cmd='''
        echo "🛑 Initiating graceful shutdown..."
        echo "1. Stopping frontend (BFF layer)..."
        kubectl scale deployment frontend -n todo-app --replicas=0 2>/dev/null || true
        sleep 5
        echo "2. Stopping backend API..."
        kubectl scale deployment backend -n todo-app --replicas=0 2>/dev/null || true
        sleep 10  # Extra time for database connections to close
        echo "3. Database will be stopped by Tilt..."
        echo "✅ Graceful shutdown prepared"
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['debug', 'shutdown'],
)

# =============================================================================
# STARTUP MESSAGE
# =============================================================================

print("""
🚀 Todo App Development Environment (BFF Architecture)

SERVICES:
┌─────────────────────────────────────────────────┐
│ Frontend (BFF): http://localhost:3000           │
│ Backend API:    http://localhost:8080 (internal)│
│ Database:       localhost:3306                  │
└─────────────────────────────────────────────────┘

ARCHITECTURE:
• Frontend serves as Backend-for-Frontend (BFF)
• API calls go through Next.js API routes (/api/*)
• Backend is not exposed externally

USEFUL COMMANDS:
• Run tests:        tilt trigger backend-test
• Frontend lint:    tilt trigger frontend-lint
• MySQL shell:      tilt trigger mysql-shell
• Health check:     tilt trigger health-check
• Generate protos:  tilt trigger protobuf-generate
• Graceful stop:    tilt trigger graceful-shutdown (before tilt down)

TIPS:
• Frontend hot reload via Next.js
• Backend hot reload via compiled binary
• Changes to package.json trigger npm install
• Use 'tilt down' to stop all services

✅ Ready for development!
""")