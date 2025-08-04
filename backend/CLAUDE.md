Instructions for Claude

## Buf Schema Registry Dependencies

```bash
go get buf.build/gen/go/wcygan/todo/connectrpc/go@latest
go get buf.build/gen/go/wcygan/todo/protocolbuffers/go@latest
```

## Recommended Test Commands

**For daily development (fast):**
```bash
go test -short -timeout 2m ./...
```

**For full test suite (with database tests):**
```bash
go test -timeout 10m ./...
```

**For CI/CD:**
```bash
go test -timeout 15m -v ./...
```

**Quick unit tests only:**
```bash
go test -timeout 1m ./internal/...
```