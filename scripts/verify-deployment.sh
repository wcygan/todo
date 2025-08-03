#!/bin/bash
echo "🔍 Verifying todo app deployment..."

echo "1. Checking Tilt status..."
curl -s http://localhost:10350 > /dev/null && echo "✅ Tilt UI accessible" || echo "❌ Tilt UI not accessible"

echo "2. Checking namespaces..."
kubectl get ns todo-app > /dev/null 2>&1 && echo "✅ todo-app namespace exists" || echo "❌ todo-app namespace missing"

echo "3. Checking pods..."
PODS=$(kubectl get pods -n todo-app --no-headers 2>/dev/null | wc -l)
READY=$(kubectl get pods -n todo-app --no-headers 2>/dev/null | grep "1/1.*Running" | wc -l)
echo "📊 Pods: $READY/$PODS ready"

echo "4. Checking services..."
kubectl get svc -n todo-app > /dev/null 2>&1 && echo "✅ Services deployed" || echo "❌ No services found"

echo "5. Testing connectivity..."
kubectl port-forward -n todo-app svc/backend-service 8080:8080 > /dev/null 2>&1 &
PF_PID=$!
sleep 2
curl -s http://localhost:8080/health > /dev/null && echo "✅ Backend responsive" || echo "❌ Backend not responding"
kill $PF_PID 2>/dev/null

echo "✅ Verification complete!"