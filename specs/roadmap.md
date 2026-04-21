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

## Phase 12 — Polish & Hardening **[DONE]**
- Rate limiting middleware via go-chi/httprate (100 req/min global, 20 req/min on /auth)
- Global error.tsx and not-found.tsx pages
- Reusable Skeleton + CardSkeleton components; jobs page now uses skeletons
- SEO metadata: title template, description, keywords, Open Graph, Twitter Card, robots
- Responsive audit + per-route error boundaries deferred to follow-up

## Phase 13 — Messaging Enhancements **[DONE]**
- Typing indicators: throttled `typing` ws message from frontend, "<name> is typing…" UI in chat thread
- Message search: `GET /api/v1/conversations/search?q=` (case-insensitive ILIKE) with search bar in conversation list
- Read times API: `GET /api/v1/conversations/{id}/read-times` returns last_read_at per participant
- Media uploads (MinIO/S3) deferred — needs object-storage infrastructure

## Phase 14 — Testing & CI **[DONE]**
- Go: handler/service/worker/provider/ws/job match unit tests pass with `-race`
- Frontend: Vitest + RTL — 37 tests across achievement card, conversation list, reaction bar, job card, hooks
- GitHub Actions workflow at `.github/workflows/ci.yml`:
  - **server** job: vet + build + race tests, with Postgres 16 and Redis 7 services
  - **web** job: typecheck + lint + vitest + Next.js build
- Stale tests revived to match current UI labels and API shapes
- Playwright e2e deferred — current Vitest coverage exercises core flows

---

# V2 — Production & Growth

Phases 1-14 shipped an end-to-end product. V2 turns it into something trustworthy, scalable, and commercially viable.

---

## Tier 1 — Trust & moat

These make Praxis meaningfully different from LinkedIn. Do these first.

### Phase 15 — Smart feed (AI digest) **[DONE]**
Adds a second tab on the home feed that runs the raw feed through Claude and returns a human-readable digest.
- `GET /api/v1/feed/smart?refresh=true` — fetches last 50 feed items (same scope as raw: self + followed) and returns `{summary, groups[], sourceCount, generatedAt, cached}`
- Redis-cached per user with configurable TTL (default 15 min) to control cost
- Prompt caching on the system instructions (Anthropic ephemeral cache, 5-min TTL) so the instruction block is billed at ~10% on repeat calls
- Config: `ANTHROPIC_API_KEY`, `SMART_FEED_MODEL` (default `claude-haiku-4-5`), `SMART_FEED_CACHE_TTL`
- Tight per-IP rate limit of 10 req/min on the endpoint — each miss costs real LLM tokens
- Frontend: Raw / Smart tab switcher on `/feed`; SmartFeed component with loading skeleton, error states (incl. graceful 503 "not configured"), summary card with gradient + sparkle badge, theme-group chips, and a Regenerate button
- Returns the real model and token-usage report in structured server logs for cost monitoring

### Phase 16 — Cryptographic verification **[NEXT]**
Currently `verification_hash` is a nullable column; the "Verified" badge is cosmetic. Phase 15 turns it into a real cryptographic attestation.
- Ed25519 server keypair stored in env/KMS; public key exposed at `/.well-known/praxis-pubkey.json`
- Signed payload: `{userId, type, sourceId, occurredAt, proofUrl}` canonicalized (RFC 8785 JCS) → SHA256 → Ed25519
- `verification_hash` and `verification_signature` columns populated at achievement-create time
- `GET /api/v1/achievements/{id}/verify` returns the payload + signature for independent verification
- "Verified" badge on cards links to a public verify page showing the signed payload and a "Verified ✓" result after client-side signature check

### Phase 16 — More achievement providers
Each new provider is ~1 week of work using the existing `AchievementProvider` interface.
- **LeetCode** — problems solved, contest rating milestones
- **Stack Overflow** — reputation milestones, top-answer tags
- **Medium / dev.to** — published articles (title, claps/reads)
- **npm / PyPI** — packages published, weekly download milestones
- **Hugging Face** — models/datasets published, download counts

---

## Tier 2 — Deferred-from-V1 polish

Items we explicitly deferred from earlier phases.

### Phase 17 — Responsive + accessibility audit
- Mobile (375w) and tablet (768w) pass on every route, prioritising chat, profile, and jobs
- Per-route React `error.tsx` boundaries (currently only global one)
- Keyboard navigation and focus management across modals
- ARIA labels on icon-only buttons
- WCAG AA colour-contrast sweep

### Phase 18 — Media uploads (MinIO/S3)
MinIO already runs in docker-compose but is unused.
- Avatar upload on profile (currently only pulled from GitHub)
- Image attachments in messages (schema: `message_attachments` table)
- Achievement proof attachments (screenshots, PDFs)
- Pre-signed URL flow: backend issues PUT URL, client uploads directly, notifies backend on success
- Image processing worker task: thumbnail generation

### Phase 19 — Transactional email
`email:notify` task already exists but only logs.
- Wire to Resend (or SES)
- Templates: welcome, new-follower, job-match, weekly digest
- Email preferences UI wired to backend flags (settings page has a placeholder)
- Unsubscribe token + public unsubscribe endpoint

### Phase 20 — Playwright e2e
Cover the flows where Vitest falls short.
- **Auth → provider connect → first sync → see achievements** (happy path)
- **Send message → other user sees typing → read receipt** (two-browser test)
- **Apply to job → match score shown → application submitted** (qualified and unqualified cases)
- Run in CI on PR merge to main

---

## Tier 3 — Product surface area

Make the job board side of the product real.

### Phase 21 — Company pages
- `/company/{slug}` public page with logo, bio, open jobs, verified employees
- Employee verification: users whose email domain matches are offered a "Confirm you work at X" button
- Company-admin role — creator of the company page becomes admin
- Company search

### Phase 22 — Hiring workflow
- Application status lifecycle: PENDING → REVIEWED → ACCEPTED | REJECTED (schema already has this)
- Recruiter dashboard: list applicants, filter by match score / location / availability
- Message-the-applicant from application detail
- Candidate-facing status tracker in `/profile` → Applications tab

### Phase 23 — Analytics dashboards
- **Personal**: achievement growth over time, profile views, match-rate across applied jobs, skill gap vs. saved jobs
- **Company**: applicant funnel, time-to-hire, which achievement types correlate with accepted offers
- Chart library: Recharts (lightweight, Tailwind-friendly)

---

## Tier 4 — Scale & operability

Things you need before inviting real users.

### Phase 24 — Observability
- OpenTelemetry tracing: API → worker → DB, exported to Jaeger locally and Honeycomb/Tempo in prod
- Structured request IDs end-to-end (already partial via chi RequestID)
- Sentry (or equivalent) for server panics and frontend errors
- Metrics dashboard: queue depth, ws connection count, DB pool stats, p50/p99 latency per route

### Phase 25 — Performance pass
- Audit slow queries; add indexes on `achievements.user_id`, `follows.(follower_id, following_id)`, `messages.conversation_id` if missing
- Eliminate N+1: provider list per profile, conversation participants, reactions per achievement
- Edge cache public profile + feed pages for anonymous viewers
- Bundle analysis + code-split heavy pages (Recharts, WebSocket client)
- Lighthouse pass — target 90+ on mobile

### Phase 26 — Production deploy
- Multi-stage Dockerfiles for `server` and `worker` (currently dev-only via Air)
- Deploy target: Vercel for web, Fly.io / Render for server + worker, Neon for Postgres, Upstash for Redis
- Secrets via platform secret store — no `.env` in the repo beyond `.example`
- Domain + TLS + CDN
- Runbook: how to deploy, rollback, rotate secrets, read logs
- Staging environment

---

## Tier 5 — Differentiators

Later-stage bets. Only after the platform has real users.

### Phase 27 — Mobile app
- Expo (React Native) reusing the existing API and types package
- Push notifications (APNs + FCM)
- Offline-first feed via react-query persistence
- Deep links to achievements, profiles, jobs

### Phase 28 — Mentor / advisor marketplace
- Bookable 1:1 sessions on your profile — calendar integration (Cal.com)
- Stripe Connect for payments (creator gets paid, platform takes fee)
- Session reviews → become an achievement type

### Phase 29 — On-chain attestation
- Optionally mint each verified achievement as an attestation (EAS on Base — cheapest / easiest)
- Users link a wallet in settings; minting is opt-in per achievement
- Third parties can verify on-chain without hitting our API at all (true decentralisation story)
- ENS support for profile URLs

---

# Cross-cutting backlog (not phase-scoped)

Things that don't warrant their own phase but need to happen over time.
- **Security**: CSP headers, CSRF tokens, SQLi review, secret scanning in CI
- **i18n**: messages to a translation layer (English only today)
- **Dark mode toggle**: hardcoded dark today; some users want a light UI
- **Admin tooling**: user support queries, achievement moderation, connector deletion
- **Data export**: user can download a JSON of everything about them (GDPR)
- **Account deletion**: real tombstone flow, not just a disabled flag
