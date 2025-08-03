# CRUSH.md

Project scaffolding for agentic coding. Commands and style for this repo.

Build/Lint/Test
- Tilt (k8s dev): tilt up | tilt down | tilt trigger protobuf-gen | tilt trigger backend-test | tilt trigger frontend-test | tilt logs <resource>
- Proto (root): buf generate | buf push
- Backend (backend/): air | go run ./cmd/server | go build -o server ./cmd/server
- Backend tests: go test ./... | go test ./... -v -race -cover | go test ./internal/handler -run TestName | go test ./... -run 'Regex'
- Frontend (frontend/): bun dev | bun run build | bun run lint | bunx tsc --noEmit
- Frontend single test: N/A (no test runner configured)

Code Style
- Imports: Go—std first, then external, then internal; TypeScript—absolute from src where configured, otherwise relative; group and sort consistently.
- Formatting: Go—gofmt/goimports; TS—ESLint (eslint-config-next) + Prettier defaults via Next.
- Types: Prefer explicit types; no any; use Zod for validation; protobuf-generated types for RPC.
- Naming: camelCase for vars/functions; PascalCase for types/components; CONSTANT_CASE for constants; Go follows Effective Go naming.
- Errors: Backend use custom errors in internal/errors with connect code mapping; wrap with context; no panics in request path.
- Concurrency: Backend store uses sync.RWMutex; respect context cancellation/timeouts.
- React/Next: App Router, Client components only when needed; shadcn/ui; Tailwind per frontend/design.md; accessibility WCAG AA.
- State: TanStack Query for server data; React Hook Form + Zod for forms.
- API: Protocol-first via buf.build/wcygan/todo; update schemas then bump generated deps per root CLAUDE.md.

Single-File Tips
- Run one Go package test: go test ./internal/store -run TestName -v
- Run grep-like focused tests: go test ./... -run 'Handler|Service'

Rules References
- CLAUDE.md files to read first: ./CLAUDE.md, backend/CLAUDE.md, frontend/CLAUDE.md, proto/CLAUDE.md
- Cursor/Copilot rules: none found
