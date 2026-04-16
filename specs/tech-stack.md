# Tech Stack

Praxis is a full-stack application with a Go API server and a Next.js frontend. The backend handles business logic, auth, and real-time communication; the frontend renders a rich SPA with server-side rendering where it matters.

## Core

| Layer | Choice | Rationale |
|---|---|---|
| Frontend framework | **Next.js 16** (App Router) | SSR + client components, file-based routing, Turbopack for fast dev |
| Language (frontend) | TypeScript (strict) | Type safety end-to-end |
| Styling | **Tailwind CSS v4** + shadcn/ui | Utility-first CSS, consistent design system, zero runtime |
| Server state | **TanStack Query** | Caching, background refetching, optimistic updates |
| Backend framework | **Go 1.22+** with Chi router | Fast, type-safe, minimal dependencies, excellent concurrency |
| Database | **PostgreSQL 16** | Relational data with JSONB for flexible fields (achievement metadata, job requirements) |
| Cache / Queue | **Redis 7** | Session store, token blacklist, job queue backend, WebSocket pub/sub |
| Real-time | **gorilla/websocket** | WebSocket hub pattern for messaging and notifications |
| Auth | JWT (access + refresh) | Stateless auth with Redis-backed token invalidation |

## Frontend Detail

- **Next.js 16** with App Router — `use(params)` pattern for async params, React 19
- **shadcn/ui** with Base UI primitives (`@base-ui/react`) — accessible, unstyled components customized with Tailwind
- **TanStack Query** — all API calls go through a typed `ApiClient` wrapper; query keys are user-scoped to prevent cache pollution
- **Custom WebSocket client** — singleton `SocketClient` with auto-reconnect, event routing via handler map; `useSocket` hook at layout level for persistent connection

## Backend Detail

- **Chi router** — lightweight, composable middleware, stdlib-compatible
- **sqlc** for type-safe SQL — write SQL, generate Go code, no ORM overhead
- **golang-migrate** for database migrations — plain SQL files, up/down
- **Zap** for structured logging — JSON in production, pretty-print in dev
- **WebSocket Hub** — goroutine-based hub with per-user client maps, broadcast channels, ping/pong keepalive

## Data

- **PostgreSQL** — primary data store for users, achievements, conversations, jobs, notifications
- **Redis** — JWT refresh token storage, rate limiting, background job queue (Asynq)
- Schema supports JSONB for achievement metadata and job requirements, enabling flexible matching without schema changes

## Infrastructure

- **Docker Compose** — PostgreSQL + Redis for local development
- **Air** — Go hot reload (watches `.go`, `.toml`, `.yaml`, `.sql` files)
- **Makefile** — `make dev` starts everything, `make test`, `make lint`, `make build`

## Provider Architecture

Achievement providers implement a common interface:

```go
type AchievementProvider interface {
    Name() string
    FetchAchievements(ctx context.Context, token string, username string) ([]Achievement, error)
}
```

A `ProviderRegistry` maps provider names to implementations. Currently GitHub is implemented; YouTube and others follow the same pattern.

## Testing (Planned)

- **Go**: testify for assertions, table-driven tests, mock repositories
- **Frontend**: Vitest + React Testing Library for component tests
- **E2E**: Playwright for full user flows
- **CI**: GitHub Actions (lint + test + build on every PR)

## What We Are Not Using

- No GraphQL — REST is sufficient at this scale and keeps the API simple
- No ORM — sqlc generates Go from SQL; we keep full control over queries
- No Docker in production yet — local development only for now
- No blockchain yet — schema has a `verification_hash` column ready for Phase 2
- No React Native — web-first; mobile is a future consideration
