# Production Architecture Decision: Next.js BFF Pattern

## Executive Summary

**Decision**: Use Next.js with API Routes as a Backend for Frontend (BFF) proxy layer.

**Key Points**:
- Single external deployment (Next.js only)
- Backend remains internal to Kubernetes cluster
- Enhanced security and simplified client configuration
- Industry-standard approach for modern web applications

## Architecture Overview

```
Internet → Ingress → Next.js (Frontend + API Routes) → Backend Service (Internal)
                           ↓
                     Browser/Client
```

## Why This Architecture?

### 1. **Security Benefits**
- Backend API never exposed to internet
- API keys and secrets stay server-side
- Single point for authentication/authorization
- Protection against direct backend attacks

### 2. **Simplified Client Configuration**
- No CORS issues
- No need for environment-specific API URLs
- Browser always calls relative `/api/*` paths
- Works seamlessly in all environments

### 3. **Better Developer Experience**
- Type-safe API proxy with shared types
- Centralized error handling
- Request/response transformation capabilities
- Easy local development

### 4. **Production Benefits**
- Single deployment to manage
- Reduced attack surface
- Better caching opportunities
- Simplified SSL/TLS configuration

## Implementation Details

### API Route Structure
```
/frontend/src/app/api/
└── [...path]/
    └── route.ts    # Catch-all proxy handler
```

### Client Configuration
```typescript
// Simple client - always uses /api
const transport = createConnectTransport({
  baseUrl: '/api',
});
```

### Kubernetes Services
```yaml
# Frontend Service (LoadBalancer/Ingress)
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
spec:
  type: LoadBalancer  # Exposed to internet
  ports:
  - port: 3000

# Backend Service (ClusterIP)
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  type: ClusterIP  # Internal only
  ports:
  - port: 8080
```

## Alternative Architectures Considered

### 1. Separate Frontend/Backend Deployments
**Pros**: Independent scaling, clear separation
**Cons**: CORS complexity, client configuration issues, both services exposed

### 2. Direct Backend Exposure
**Pros**: Simpler initial setup
**Cons**: Security risks, CORS issues, client needs backend URL

### 3. API Gateway (Kong, Traefik)
**Pros**: Advanced routing, rate limiting
**Cons**: Additional complexity, another service to manage

## Industry Best Practices Alignment

This architecture follows established patterns:

1. **Backend for Frontend (BFF)** - Recommended by ThoughtWorks, Microsoft
2. **API Routes as Proxy** - Official Next.js pattern
3. **Zero Trust Security** - Backend not exposed to internet
4. **12-Factor App** - Clear separation of concerns

## Migration Path

### Current State
- Frontend tries to connect directly to backend
- Browser can't resolve Kubernetes service names
- Configuration complexity

### Target State
1. Frontend calls `/api/*` routes
2. API routes proxy to backend service
3. Single external endpoint

### Steps
1. ✅ Create API route proxy handler
2. ✅ Update client to use `/api` base URL
3. ✅ Update Kubernetes configmap
4. ⏳ Rebuild and deploy frontend image
5. ⏳ Verify end-to-end functionality

## Common Concerns Addressed

### Q: What about microservices independence?
A: The backend remains a separate service. Only the API gateway functionality is integrated with the frontend.

### Q: Performance overhead?
A: Minimal - Next.js API routes run on the same server. The proxy adds microseconds, not meaningful latency.

### Q: What if we need to expose the API publicly later?
A: The backend can still be exposed separately if needed. The BFF pattern doesn't prevent this.

### Q: How do we handle multiple backends?
A: The API route can intelligently route to different backend services based on the path.

## Conclusion

The Next.js BFF pattern provides the best balance of:
- **Security**: Backend stays internal
- **Simplicity**: Single deployment, no CORS
- **Flexibility**: Can evolve as needs change
- **Industry alignment**: Follows proven patterns

This is the recommended approach for modern Kubernetes-deployed web applications.