# Todo Backend Test Suite

Comprehensive test suite for the Todo Backend + MariaDB system with multiple test levels and failure scenarios.

## Test Organization

```
test/
├── unit/                    # Unit tests with isolated components
│   ├── store_manager_test.go      # Store manager initialization tests
│   └── mysql_store_unit_test.go   # MySQL store CRUD operations
├── integration/             # Integration tests with full stack
│   ├── full_stack_test.go         # Complete backend + DB testing
│   ├── performance_test.go        # Performance and load testing
│   └── failure_scenarios_test.go  # Database resilience tests
└── e2e/                    # End-to-end tests
    └── kubernetes_e2e_test.go     # Kubernetes deployment testing
```

## Test Types

### Unit Tests (`test/unit/`)
- **store_manager_test.go**: Database connection management and configuration validation
- **mysql_store_unit_test.go**: CRUD operations, context handling, edge cases, Unicode support

### Integration Tests (`test/integration/`)
- **full_stack_test.go**: Complete HTTP/ConnectRPC stack with MariaDB using testcontainers
- **performance_test.go**: Throughput, latency, concurrent operations, mixed workload testing
- **failure_scenarios_test.go**: Database connection failures, transaction integrity, resource exhaustion

### E2E Tests (`test/e2e/`)
- **kubernetes_e2e_test.go**: Real deployment testing with configurable backend URLs

## Running Tests

### Quick Tests
```bash
go test ./test/... -short
```

### Full Test Suite
```bash
# All tests (requires Docker for testcontainers)
go test ./test/...

# Specific test categories
go test ./test/unit/...
go test ./test/integration/...
go test ./test/e2e/...
```

### E2E Tests
```bash
# Configure backend URL (default: http://localhost:8080)
export E2E_BACKEND_URL="http://your-backend:8080"
export E2E_TIMEOUT="60s"
export E2E_SKIP_CLEANUP="false"

go test ./test/e2e/...
```

## Test Features

### Comprehensive Coverage
- **CRUD Operations**: Create, Read, Update, Delete with validation
- **Concurrent Operations**: Multi-threaded safety and race condition testing
- **Error Handling**: Invalid inputs, non-existent resources, network failures
- **Performance**: Throughput benchmarks and latency measurements
- **Resilience**: Database failures, connection recovery, resource exhaustion

### Advanced Scenarios
- **Unicode Support**: International characters and emojis
- **Large Data**: 10KB+ task descriptions and bulk operations
- **Memory Pressure**: 5000+ tasks with memory usage monitoring
- **Context Handling**: Cancellation and timeout behavior
- **Connection Pooling**: Pool exhaustion and recovery testing

### Technology Stack
- **TestContainers**: Isolated MariaDB 11.5 instances per test
- **ConnectRPC**: Full protocol testing with HTTP/2 support
- **Testify**: Assertions and test structuring
- **Context**: Proper cancellation and timeout handling

## Performance Baselines

### Expected Throughput
- **Task Creation**: >100 tasks/sec (sequential), >200 tasks/sec (concurrent)
- **Task Retrieval**: >500 retrievals/sec
- **Task Listing**: >50 listings/sec with 500+ tasks
- **Mixed Operations**: >100 ops/sec with <1% error rate

### Latency Targets
- **Individual Operations**: <200ms average
- **Bulk Operations**: <2s for 1000 tasks
- **Large Tasks**: <5s creation for 50KB descriptions

## Test Data Cleanup

All tests include proper cleanup mechanisms:
- Automatic container termination
- Task deletion after test completion
- Resource leak prevention
- Graceful test interruption handling