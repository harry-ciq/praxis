# Roadmap

Phases are shippable slices — each one adds a working feature that can be tested independently.

---

## Phase 1 — Monorepo Scaffold + Infra **[DONE]**
- Monorepo structure (`web/`, `server/`, `docker-compose.yml`, `Makefile`)
- Docker Compose: PostgreSQL 16, Redis 7
- Go module + Chi router + health endpoint
- Next.js 16 + Tailwind v4 + shadcn/ui
- Air hot reload for Go, Turbopack for Next.js
- `make dev` starts the full stack

## Phase 2 — Database + Auth **[DONE]**
- SQL migrations for all core tables
- JWT-based auth (access + refresh tokens, Redis-backed)
- GitHub OAuth: Go handles callback, creates user, returns JWT
- Auth middleware (protected + optional-auth routes)
- Next.js auth context with automatic token refresh
- Sign-in / sign-up pages

## Phase 3 — GitHub Integration + Achievement Engine **[DONE]**
- GitHub provider implementing the `AchievementProvider` interface
- Achievement detection: repo creation, star milestones, commit streaks, PR merges, first OSS contribution
- Deduplication via `source_id` (no duplicate achievements)
- `POST /api/v1/achievements/sync` triggers a sync
- Auto-sync on login

## Phase 4 — Feed + Achievement Cards **[DONE]**
- Feed endpoint with cursor pagination (`GET /api/v1/feed`)
- Feed scoped to self + followed users (not global)
- Achievement cards with type icons, metadata, proof links, timestamps
- Reaction endpoints (add/remove) with optimistic UI updates
- User-scoped query cache to prevent stale data across sessions

## Phase 5 — User Profiles + Follows **[DONE]**
- Profile page with achievements, stats, experiences, skills
- Follow / unfollow with cache invalidation (feed refreshes on follow change)
- User search & discovery with prefix-priority ranking
- Inline follow buttons in search results
- Settings page for profile editing and connected providers

## Phase 6 — Real-Time Messaging **[DONE]**
- WebSocket hub (gorilla/websocket) with per-user client tracking
- Conversation CRUD: create, list, delete
- Message send / receive / delete with real-time WebSocket broadcast
- Messaging UI: conversation list, chat thread, message bubbles
- New conversation flow with user search
- Socket client with exponential-backoff reconnection
- Delete message (own only) and delete conversation

## Phase 7 — Job Board **[DONE]**
- Job endpoints: list, detail, create, apply
- Job listing page + job detail page
- Basic application submission

## Phase 8 — Notifications **[DONE]**
- Notification service + repository
- Real-time WebSocket delivery
- Notification bell + dropdown + dedicated page
- List, unread count, mark-as-read

---

## Phase 9 — Achievement-Based Job Matching **[DONE]**
- Threshold-aware match scoring: `TYPE` (any), `TYPE:N` for thresholds (stars/subs/views) or counts (PRs/repos/videos)
- `?qualified=true` filter returns only fully qualified jobs
- Match badge with % on job cards (green at 100%, amber partial, grey 0)
- Per-requirement check marks on job detail page
- Nested company shape in job response

## Phase 10 — Background Workers **[DONE]**
- Asynq (Redis-backed) job queue with critical/default/low queues
- `provider:sync` task — re-syncs achievements for any connected provider
- Periodic scheduler enqueues syncs every `SYNC_INTERVAL_HOURS` (default 6h)
- `POST /api/v1/webhooks/github` with HMAC-SHA256 signature verification, enqueues a critical-priority sync for the sender
- `email:notify` task placeholder (logs only; ready for SES/SendGrid wiring)

## Phase 11 — Additional Providers **[DONE]**
- YouTube provider: video published, subscriber milestones, view milestones, channel created
- Google OAuth flow for connecting YouTube accounts (separate from GitHub OAuth)
- YouTube connect UI on settings page with OAuth redirect flow
- Achievement cards with distinct icons for each YouTube achievement type
- Provider registry is already extensible — add new providers by implementing the interface
- Per-provider sync settings in user profile

## Phase 12 — Polish & Hardening **[NOT STARTED]**
- Responsive design audit (mobile + tablet)
- Error boundaries on all routes
- Loading skeletons instead of spinners
- 404 / 500 error pages
- Input sanitization on all forms
- Rate limiting middleware
- SEO meta tags + Open Graph images

## Phase 13 — Messaging Enhancements **[NOT STARTED]**
- Typing indicators (server-side handling exists, needs frontend UI)
- Read receipts (last_read_at tracked, needs visual indicator)
- Message search
- Media uploads (images, files) via MinIO/S3

## Phase 14 — Testing & CI **[NOT STARTED]**
- Go unit tests with testify (services, handlers)
- Vitest component tests (React Testing Library)
- Playwright e2e tests (auth flow, feed, messaging)
- GitHub Actions CI pipeline (lint, test, build)

---

Later phases (not yet planned): blockchain verification of achievements, company pages, therapist-style profile pages for mentors, reporting/analytics dashboard, mobile app.
