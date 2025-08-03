#!/bin/bash
echo "🔍 Verifying todo app deployment using Tilt CLI..."

echo "1. Checking Tilt status..."
if pgrep -f "tilt" > /dev/null; then
    echo "✅ Tilt process running"
else
    echo "❌ Tilt not running"
    exit 1
fi

echo "2. Getting all Tilt resources..."
tilt get all

echo "3. Describing resource details..."
echo "--- Frontend ---"
tilt describe frontend 2>/dev/null || echo "❌ Frontend resource not found"

echo "--- Backend ---"
tilt describe backend 2>/dev/null || echo "❌ Backend resource not found"

echo "4. Checking resource logs (last 10 lines)..."
echo "--- Frontend Logs ---"
tilt logs frontend --tail=10 2>/dev/null || echo "❌ No frontend logs available"

echo "--- Backend Logs ---"
tilt logs backend --tail=10 2>/dev/null || echo "❌ No backend logs available"

echo "5. Tilt environment diagnostics..."
tilt doctor

echo "6. Kubernetes resources via Tilt context..."
kubectl get pods -n todo-app 2>/dev/null || echo "❌ No pods found in todo-app namespace"

echo "7. Service connectivity test..."
kubectl port-forward -n todo-app svc/backend-service 8080:8080 > /dev/null 2>&1 &
PF_PID=$!
sleep 2
curl -s http://localhost:8080/health > /dev/null && echo "✅ Backend health check passed" || echo "❌ Backend health check failed"
kill $PF_PID 2>/dev/null

kubectl port-forward -n todo-app svc/frontend-service 3000:3000 > /dev/null 2>&1 &
PF_PID=$!
sleep 2
curl -s http://localhost:3000 > /dev/null && echo "✅ Frontend accessible" || echo "❌ Frontend not accessible"
kill $PF_PID 2>/dev/null

echo "✅ Tilt CLI verification complete!"
echo ""
echo "💡 Additional Tilt commands you can use:"
echo "  tilt get resources          # List all resources"
echo "  tilt describe <resource>    # Detailed resource info"
echo "  tilt logs <resource>        # Stream resource logs"
echo "  tilt wait --for=condition=Ready resource/<name>  # Wait for resource to be ready"
echo "  tilt dump engine           # Dump internal Tilt state for debugging"