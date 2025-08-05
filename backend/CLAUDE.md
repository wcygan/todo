Instructions for Claude

## Buf Schema Registry Dependencies

```bash
go get buf.build/gen/go/wcygan/todo/connectrpc/go@latest
go get buf.build/gen/go/wcygan/todo/protocolbuffers/go@latest
```

## Recommended Test Commands

**For daily development (fast, parallel, verbose):**
```bash
go test -v -short -timeout 30s -parallel 8 ./...
```

**For full test suite (with database tests):**
```bash
go test -v -timeout 2m -parallel 8 ./...
```

**For CI/CD:**
```bash
go test -v -timeout 5m -parallel 8 ./...
```

**Quick unit tests only:**
```bash
go test -v -short -timeout 30s -parallel 8 ./internal/...
```

**Run with test result caching:**
```bash
# Use cached results for unchanged code (with progress)
go test -v -timeout 30s -parallel 8 ./...

# Force fresh run (no cache)
go test -v -count=1 -timeout 30s -parallel 8 ./...
```

**Watch test output format:**
```bash
# Minimal output (fast)
go test -short -timeout 30s -parallel 8 ./...

# Verbose output (see each test)
go test -v -short -timeout 30s -parallel 8 ./...

# JSON output (for tooling)
go test -json -short -timeout 30s -parallel 8 ./...
```

## Test Optimization Guidelines

### 1. Run Tests in Parallel
Add `t.Parallel()` as the first line in test functions:
```go
func TestExample(t *testing.T) {
    t.Parallel()
    // test code...
}
```

### 2. Use Test Caching
Go caches test results automatically. Only use `go clean -testcache` when necessary.

### 3. Skip Slow Tests During Development
```bash
# Skip integration tests
go test -short ./...
```

### 4. Run Specific Tests
```bash
# Run only tests matching pattern
go test -run TestTaskCreation ./...

# Run specific package
go test ./internal/handler
```