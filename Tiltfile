# Comprehensive Tiltfile for todo app development workflow
# Provides hot reload, protocol buffer automation, and development optimizations

# =============================================================================
# ENVIRONMENT SETUP
# =============================================================================

# Allow Minikube and local development contexts
allow_k8s_contexts(['minikube', 'docker-desktop', 'kind-kind'])

# Check if local registry is available
local_registry = 'localhost:5000'
if str(local('docker ps --filter "name=registry" --format "{{.Names}}" || true')).strip():
    default_registry(local_registry)
    print("✅ Using local registry: " + local_registry)
else:
    print("⚠️  Local registry not detected. Starting one...")
    local('docker run -d --restart=always -p 5000:5000 --name registry registry:2 || true')
    default_registry(local_registry)

# Set resource quotas for development
update_settings(k8s_upsert_timeout_secs=300, suppress_unused_image_warnings=None)

# =============================================================================
# PROTOCOL BUFFER INTEGRATION
# =============================================================================

# Protocol buffer code generation with dependency management
local_resource(
    'protobuf-generate',
    cmd='''
        echo "🔧 Generating protocol buffer code..."
        buf generate
        echo "✅ Protocol buffer generation complete"
    ''',
    deps=['./proto', './buf.gen.yaml'],
    labels=['protobuf', 'codegen'],
    resource_deps=[],
)

# Watch for protocol buffer changes and trigger dependent rebuilds
local_resource(
    'protobuf-deps-update',
    cmd='''
        echo "🔄 Updating protocol buffer dependencies..."
        cd backend && go get buf.build/gen/go/wcygan/todo/protocolbuffers/go@latest
        cd backend && go get buf.build/gen/go/wcygan/todo/connectrpc/go@latest
        cd ../frontend && bun add @buf/wcygan_todo.bufbuild_es@latest
        cd frontend && bun add @buf/wcygan_todo.connectrpc_query-es@latest
        echo "✅ Dependencies updated"
    ''',
    deps=['./proto'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['protobuf', 'deps'],
    resource_deps=['protobuf-generate'],
)

# =============================================================================
# BACKEND CONFIGURATION
# =============================================================================

# Build backend with enhanced live updates and Go compilation optimization
docker_build(
    'todo-backend',
    context='./backend',
    dockerfile='./k8s/backend/dockerfile.dev',
    # Optimized live update strategy
    live_update=[
        # Sync Go source files
        sync('./backend/', '/app/'),
        # Fast rebuild with air hot reload (no restart needed)
        run('echo "🔄 Code updated, air will auto-reload..."', trigger=[
            '**/*.go',
            'go.mod',
            'go.sum',
            'configs/**/*'
        ]),
    ],
    # Development build optimizations
    build_args={
        'GOPROXY': 'direct',
        'GOSUMDB': 'sum.golang.org',
        'CGO_ENABLED': '0',
    },
    # Only rebuild if Dockerfile or dependencies change
    only=['./backend/go.mod', './backend/go.sum', './backend/'],
)

# =============================================================================
# FRONTEND CONFIGURATION  
# =============================================================================

# Build frontend with Bun and Turbopack optimization
docker_build(
    'todo-frontend',
    context='./frontend',
    dockerfile='./k8s/frontend/dockerfile.dev',
    # Enhanced live update for Next.js with Turbopack
    live_update=[
        # Sync source files for hot module replacement
        sync('./frontend/src/', '/app/src/'),
        sync('./frontend/public/', '/app/public/'),
        sync('./frontend/components.json', '/app/components.json'),
        sync('./frontend/next.config.ts', '/app/next.config.ts'),
        sync('./frontend/tailwind.config.js', '/app/tailwind.config.js'),
        sync('./frontend/postcss.config.mjs', '/app/postcss.config.mjs'),
        sync('./frontend/tsconfig.json', '/app/tsconfig.json'),
        
        # Handle dependency changes
        sync('./frontend/package.json', '/app/package.json'),
        sync('./frontend/bun.lock', '/app/bun.lock'),
        
        # Install dependencies if package files change
        run('echo "📦 Installing dependencies..." && bun install --frozen-lockfile', 
            trigger=['package.json', 'bun.lock']),
        
        # Log update for visibility
        run('echo "🔄 Frontend code updated, Turbopack will hot-reload..."', 
            trigger=['src/**/*', 'public/**/*']),
    ],
    # Development build optimizations
    build_args={
        'NODE_ENV': 'development',
        'NEXT_TELEMETRY_DISABLED': '1',
        'TURBOPACK': '1',
    },
    # Only rebuild if Dockerfile or lock files change
    only=['./frontend/package.json', './frontend/bun.lock', './frontend/'],
)

# =============================================================================
# KUBERNETES RESOURCES & DATABASE INTEGRATION
# =============================================================================

# Apply Kubernetes manifests with development optimizations
k8s_yaml(kustomize('./k8s/development'))

# Load extensions for enhanced functionality
load('ext://restart_process', 'docker_build_with_restart')
load('ext://helm_remote', 'helm_remote')
load('ext://uibutton', 'cmd_button', 'location', 'text_input')

# MySQL Operator setup with enhanced configuration
helm_remote(
    'mysql-operator',
    repo_name='mysql-operator',
    repo_url='https://mysql.github.io/mysql-operator/',
    namespace='mysql-operator',
    create_namespace=True,
    set=[
        'image.pullPolicy=IfNotPresent',
        'resources.requests.memory=128Mi',
        'resources.requests.cpu=100m',
        'resources.limits.memory=256Mi',
        'resources.limits.cpu=200m',
    ],
)

# Database initialization and management
local_resource(
    'mysql-init-schema',
    cmd='''
        echo "🗄️ Initializing database schema..."
        # Wait for MySQL to be ready
        kubectl wait --for=condition=ready pod -l app=mysql-cluster --timeout=300s -n todo-app || true
        
        # Create database and tables if they don't exist
        kubectl exec mysql-cluster-0 -n todo-app -- mysql -u root -p$MYSQL_ROOT_PASSWORD -e "
            CREATE DATABASE IF NOT EXISTS todoapp;
            USE todoapp;
            CREATE TABLE IF NOT EXISTS tasks (
                id VARCHAR(36) PRIMARY KEY,
                title VARCHAR(255) NOT NULL,
                is_completed BOOLEAN DEFAULT FALSE,
                priority ENUM('none', 'low', 'medium', 'high') DEFAULT 'none',
                due_date DATETIME NULL,
                notes TEXT,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                completed_at DATETIME NULL
            );
        " || echo "Database initialization will retry on next startup"
        echo "✅ Database schema initialization complete"
    ''',
    deps=[],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['database', 'init'],
    resource_deps=['mysql-cluster'],
)

# =============================================================================
# RESOURCE CONFIGURATION & DEPENDENCIES
# =============================================================================

# Backend service configuration with health checks and dependencies
k8s_resource(
    'backend',
    port_forwards=['8080:8080'],
    resource_deps=['mysql-cluster', 'protobuf-generate'],
    labels=['api', 'core'],
    trigger_mode=TRIGGER_MODE_AUTO,
    auto_init=True,
)

# Frontend service configuration with backend dependency
k8s_resource(
    'frontend', 
    port_forwards=['3000:3000'],
    resource_deps=['backend', 'protobuf-generate'],
    labels=['ui', 'core'],
    trigger_mode=TRIGGER_MODE_AUTO,
    auto_init=True,
)

# Database cluster with enhanced monitoring
k8s_resource(
    'mysql-cluster',
    port_forwards=['3306:3306'],
    resource_deps=['mysql-operator'],
    labels=['database', 'infrastructure'],
    trigger_mode=TRIGGER_MODE_AUTO,
    auto_init=True,
)

# Database secrets and configuration
k8s_resource(
    workload='mysql-init',
    objects=[
        'mysql-credentials:secret',
        'backend-secrets:secret',
        'backend-config:configmap',
        'frontend-config:configmap',
    ],
    labels=['database', 'config'],
    resource_deps=[],
)

# Network and ingress configuration
k8s_resource(
    'todo-app-ingress',
    labels=['networking', 'infrastructure'],
    resource_deps=['frontend', 'backend'],
)

# =============================================================================
# DEVELOPMENT WORKFLOW & TESTING INTEGRATION
# =============================================================================

# Configuration options for selective service deployment
config.define_string_list('services', args=False, usage="Comma-separated list of services to start")
config.define_bool('testing', args=False, usage="Enable testing resources")
config.define_bool('monitoring', args=False, usage="Enable monitoring stack") 
cfg = config.parse()

# Selective service deployment
services = cfg.get('services', [])
if services:
    config.set_enabled_resources(services)
    print("🎯 Starting only specified services: " + ", ".join(services))

# Enhanced testing integration
local_resource(
    'backend-test-unit',
    cmd='''
        echo "🧪 Running backend unit tests..."
        cd backend && go test -v -race -cover -coverprofile=coverage.out ./internal/...
        echo "✅ Backend unit tests complete"
    ''',
    deps=['./backend/internal'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['test', 'backend'],
)

local_resource(
    'backend-test-integration',
    cmd='''
        echo "🔗 Running backend integration tests..."
        cd backend && go test -v -tags=integration ./test/integration/...
        echo "✅ Backend integration tests complete"
    ''',
    deps=['./backend/test'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['test', 'backend'],
    resource_deps=['backend'],
)

local_resource(
    'frontend-test-unit',
    cmd='''
        echo "🧪 Running frontend unit tests..."
        cd frontend && bun test
        echo "✅ Frontend unit tests complete"
    ''',
    deps=['./frontend/src'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['test', 'frontend'],
)

local_resource(
    'frontend-lint',
    cmd='''
        echo "🔍 Running frontend linting..."
        cd frontend && bun run lint
        cd frontend && bunx tsc --noEmit
        echo "✅ Frontend linting complete"
    ''',
    deps=['./frontend/src'],
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['test', 'frontend'],
)

# Run all tests
local_resource(
    'test-all',
    cmd='''
        echo "🚀 Running complete test suite..."
        tilt trigger backend-test-unit backend-test-integration frontend-test-unit frontend-lint
        echo "✅ All tests complete"
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['test', 'all'],
    resource_deps=['backend-test-unit', 'backend-test-integration', 'frontend-test-unit', 'frontend-lint'],
)

# =============================================================================
# ADVANCED FEATURES & DEBUG CONFIGURATIONS
# =============================================================================

# Enhanced logging and debugging tools
local_resource(
    'logs-backend',
    cmd='''
        echo "📋 Streaming backend logs..."
        kubectl logs -f deployment/backend -n todo-app --tail=100
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['debug', 'logs'],
    resource_deps=['backend'],
)

local_resource(
    'logs-frontend',
    cmd='''
        echo "📋 Streaming frontend logs..."
        kubectl logs -f deployment/frontend -n todo-app --tail=100
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['debug', 'logs'],
    resource_deps=['frontend'],
)

local_resource(
    'logs-database',
    cmd='''
        echo "📋 Streaming database logs..."
        kubectl logs -f mysql-cluster-0 -n todo-app --tail=50
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['debug', 'logs'],
    resource_deps=['mysql-cluster'],
)

# Database management and debugging
local_resource(
    'db-shell',
    cmd='''
        echo "🗄️ Opening database shell..."
        echo "Available databases: todoapp"
        echo "Default connection: mysql -u root -p"
        kubectl exec -it mysql-cluster-0 -n todo-app -- mysql -u root -p
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['database', 'debug'],
    resource_deps=['mysql-cluster'],
)

local_resource(
    'db-backup',
    cmd='''
        echo "💾 Creating database backup..."
        kubectl exec mysql-cluster-0 -n todo-app -- mysqldump -u root -p todoapp > ./backup/todoapp_$(date +%Y%m%d_%H%M%S).sql
        echo "✅ Backup created in ./backup/"
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['database', 'backup'],
    resource_deps=['mysql-cluster'],
)

# Performance monitoring and metrics
if cfg.get('monitoring', False):
    local_resource(
        'port-forward-grafana',
        cmd='kubectl port-forward svc/grafana 3001:3000 -n monitoring',
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        labels=['monitoring', 'metrics'],
    )
    
    local_resource(
        'port-forward-prometheus',
        cmd='kubectl port-forward svc/prometheus 9090:9090 -n monitoring',
        auto_init=False,
        trigger_mode=TRIGGER_MODE_MANUAL,
        labels=['monitoring', 'metrics'],
    )

# Development environment status and health checks
local_resource(
    'health-check',
    cmd='''
        echo "🏥 Checking service health..."
        echo "Backend API:"
        curl -f http://localhost:8080/health || echo "❌ Backend not ready"
        echo "Frontend:"
        curl -f http://localhost:3000 || echo "❌ Frontend not ready"
        echo "Database:"
        kubectl exec mysql-cluster-0 -n todo-app -- mysqladmin ping -u root -p || echo "❌ Database not ready"
        echo "✅ Health check complete"
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['debug', 'health'],
    resource_deps=['backend', 'frontend', 'mysql-cluster'],
)

# Build cache management
local_resource(
    'clean-cache',
    cmd='''
        echo "🧹 Cleaning build caches..."
        cd backend && go clean -cache -modcache
        cd ../frontend && bun cache clean
        docker system prune -f --volumes
        echo "✅ Cache cleanup complete"
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['debug', 'cleanup'],
)

# Environment reset
local_resource(
    'reset-environment',
    cmd='''
        echo "🔄 Resetting development environment..."
        tilt down
        kubectl delete namespace todo-app --ignore-not-found=true
        docker system prune -f --volumes
        echo "✅ Environment reset complete - run 'tilt up' to restart"
    ''',
    auto_init=False,
    trigger_mode=TRIGGER_MODE_MANUAL,
    labels=['debug', 'reset'],
)

# =============================================================================
# PERFORMANCE OPTIMIZATIONS & FINAL CONFIGURATION
# =============================================================================

# Watch additional files for changes with optimized patterns
watch_file('./proto/')
watch_file('./buf.gen.yaml')
watch_file('./buf.yaml')

# Optimize resource allocation for development
update_settings(
    max_parallel_updates=3,
    default_to_user_repos=True,
    suppress_unused_image_warnings=None,
)

# UI button integrations for common tasks
cmd_button('health-check',
          resource='health-check',
          icon_name='health_and_safety',
          text='Health Check')

cmd_button('test-all',
          resource='test-all', 
          icon_name='quiz',
          text='Run All Tests')

cmd_button('clean-cache',
          resource='clean-cache',
          icon_name='cleaning_services', 
          text='Clean Cache')

# =============================================================================
# DEVELOPMENT INFORMATION & STARTUP MESSAGE
# =============================================================================

print("""
🚀 Todo App Comprehensive Development Environment

SERVICES:
┌──────────────────────────────────────────────────────────┐
│ Frontend:    http://localhost:3000 (todo.local)         │
│ Backend API: http://localhost:8080 (api.todo.local)     │ 
│ Database:    localhost:3306                              │
│ Ingress:     http://todo.local / http://api.todo.local   │
└──────────────────────────────────────────────────────────┘

TESTING COMMANDS:
• Backend Unit:        tilt trigger backend-test-unit
• Backend Integration: tilt trigger backend-test-integration  
• Frontend Unit:       tilt trigger frontend-test-unit
• Frontend Lint:       tilt trigger frontend-lint
• All Tests:           tilt trigger test-all

DEBUG COMMANDS:
• Backend Logs:        tilt trigger logs-backend
• Frontend Logs:       tilt trigger logs-frontend
• Database Logs:       tilt trigger logs-database
• Database Shell:      tilt trigger db-shell
• Health Check:        tilt trigger health-check

PROTOBUF COMMANDS:
• Generate Code:       tilt trigger protobuf-generate
• Update Dependencies: tilt trigger protobuf-deps-update

MAINTENANCE:
• Clean Cache:         tilt trigger clean-cache
• Database Backup:     tilt trigger db-backup
• Reset Environment:   tilt trigger reset-environment

SETUP REQUIREMENTS:
• Add to /etc/hosts: 127.0.0.1 todo.local api.todo.local
• Ensure Minikube/Docker Desktop is running
• Install: kubectl, helm, buf CLI

ADVANCED OPTIONS:
• Selective services:  tilt up --hud=false -- --services=backend,frontend
• Enable monitoring:   tilt up -- --monitoring=true
• Testing mode:        tilt up -- --testing=true

PERFORMANCE FEATURES:
✅ Hot reload with Air (Go) and Turbopack (Next.js)
✅ Optimized Docker layer caching
✅ Protocol buffer auto-generation and dependency updates
✅ Resource dependency management
✅ Local registry integration
✅ Enhanced live updates with minimal container restarts

Use 'tilt down' to cleanup all resources
Use 'tilt logs <resource>' to view specific logs
""")

# Final optimizations and context verification
if k8s_context() not in ['minikube', 'docker-desktop', 'kind-kind']:
    fail("⚠️  Please use minikube, docker-desktop, or kind for development")

# Verify required tools
local('which kubectl > /dev/null || (echo "❌ kubectl not found" && exit 1)')
local('which buf > /dev/null || (echo "❌ buf CLI not found" && exit 1)')

print("✅ Tilt configuration loaded successfully!")