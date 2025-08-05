# Todo App Scripts

This directory contains Deno TypeScript scripts for development automation.

## Available Scripts

### fast-test.ts
Runs the Go test suite with optimizations for speed:
- Verbose output (`-v`) to show test progress
- Uses `-short` flag to skip integration tests
- Runs tests in parallel with `-parallel 8`
- Shows execution time
- Uses Go's test result caching

### test-progress.ts
Simple test runner with clear progress indicators:
- Shows which tests are running
- Displays timing for each test
- Color-coded output
- Clear summary at the end

### test-watch.ts
Advanced test watcher with formatted output:
- Real-time test progress
- Emoji indicators for test status
- Formatted output for better readability
(Note: Currently being refined)

### optimize-tests.ts
Adds `t.Parallel()` to test functions for concurrent execution.
(Currently not fully implemented - manual optimization recommended)

## Deno Tasks

All tasks are defined in `/deno.json`:

```bash
# Docker Management
deno task up       # Start services in background
deno task down     # Stop all services
deno task dev      # Build and start services (foreground)
deno task logs     # Follow service logs
deno task restart  # Restart running services
deno task clean    # Stop services and clean Docker system

# Testing
deno task test       # Run fast tests (skips integration)
deno task test:full  # Run complete test suite
deno task test:unit  # Run unit tests only
```

## Performance Notes

- Tests run in ~3 seconds with caching enabled
- Use `go clean -testcache` to force fresh test runs
- Integration tests with TestContainers add significant time

## Requirements

- Deno 2.0+
- Go 1.24+
- Docker
- `@david/dax` package (auto-installed)