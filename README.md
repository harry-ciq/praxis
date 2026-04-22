# Praxis

**Where execution speaks louder than words.**

Praxis is a professional network driven by verified execution, not self-promotion. Instead of users writing about what they did, Praxis connects to their GitHub, YouTube, and other platforms to automatically detect, verify, and surface real achievements -- repos created, PRs merged, commit streaks, stars milestones, and more.

Think LinkedIn, but every achievement on your profile is backed by proof.

## Why Praxis?

The problem with existing professional networks: anyone can claim anything. Praxis flips this by making the platform itself the verifier. Your profile is built from what you actually ship, not what you say you shipped.

- **Automatic achievement detection** -- connect GitHub and YouTube; Praxis tracks repos, commits, PRs, stars, video uploads, subscriber milestones, and more
- **Verified proof** -- every achievement links back to its source. Cryptographic signatures coming in Phase 16
- **Achievement-based job matching** -- employers post requirements like `COMMIT_STREAK:30` or `STARS_MILESTONE:100`, and Praxis surfaces match scores per candidate
- **Smart feed** -- a Claude-generated weekly digest with hero stat, themed chips, "watching" predictions, and near-miss milestones. Each claim links back to the underlying achievement
- **Real-time feed + messaging** -- WebSocket-driven activity stream, live conversations with typing indicators

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS v4, shadcn/ui, Playfair Display (editorial serif) |
| **Backend** | Go 1.22+, Chi router, JWT auth, go-chi/httprate (rate limiting) |
| **Database** | PostgreSQL 16 (pgx driver), Redis 7 |
| **Real-time** | WebSocket (gorilla/websocket), per-user client hub |
| **Background Jobs** | Asynq (Redis-backed task queue) with periodic scheduler |
| **External APIs** | GitHub API (go-github), YouTube Data API v3 |
| **AI / LLM** | Anthropic Messages API (raw HTTP, prompt caching) for the Smart Feed digest |
| **Object Storage** | MinIO (S3-compatible) — provisioned, not yet wired up |
| **Testing** | Go testify, Vitest + React Testing Library |
| **CI** | GitHub Actions (Postgres + Redis services, race tests, type-check, build) |

## Architecture

```
                    +-------------------------+
                    |    Next.js Frontend     |
                    |    Port 3000            |
                    +-----------+-------------+
                                |
                       REST + WebSocket
                                |
                    +-----------v-------------+        +------------------+
                    |     Go API Server       |------->| Anthropic API    |
                    |     Port 8080           |  LLM   | (Claude Messages)|
                    |                         |        +------------------+
                    |  Handlers -> Services   |
                    |  -> Repositories (pgx)  |
                    |                         |
                    |  Provider Registry      |        +------------------+
                    |   +- GitHubProvider     |------->| GitHub API       |
                    |   +- YouTubeProvider    |------->| YouTube Data API |
                    +-+--------+----------+---+        +------------------+
                      |        |          |
               +------v-+  +--v---+  +---v----+
               |Postgres |  |Redis |  | MinIO  |
               +----+----+  +--+---+  +--------+
                                |
                                |  shared queue + smart-feed cache
                          +-----v--------+
                          | Asynq Worker |
                          | - provider:sync (periodic + webhook)
                          | - email:notify
                          +--------------+
                                ^
                                |  /api/v1/webhooks/github (HMAC-validated)
                          +-----+--------+
                          | GitHub Push  |
                          +--------------+
```

## Project Structure

```
praxis/
├── docker-compose.yml        # PostgreSQL + Redis + MinIO
├── Makefile                  # Dev commands (make dev, make test, etc.)
│
├── server/                   # Go backend
│   ├── cmd/
│   │   ├── api/main.go       # HTTP + WebSocket server
│   │   └── worker/main.go    # Background job worker + periodic scheduler
│   ├── internal/
│   │   ├── config/            # Environment config (godotenv.Overload — .env wins over shell)
│   │   ├── handler/           # HTTP handlers (auth, feed, achievements, jobs, messages, webhooks)
│   │   ├── service/           # Business logic (achievements, smart-feed digest, jobs, messages)
│   │   ├── repository/        # Database access (pgx)
│   │   ├── provider/          # Achievement providers (GitHub, YouTube)
│   │   ├── llm/               # Anthropic Messages API client (raw HTTP, prompt caching)
│   │   ├── middleware/        # Auth, CORS, logging, rate limiting
│   │   ├── ws/                # WebSocket hub for real-time messaging
│   │   └── worker/            # Asynq task definitions + handlers + scheduler
│   └── migrations/            # SQL migration files
│
├── web/                      # Next.js frontend
│   └── src/
│       ├── app/               # App Router pages
│       │   ├── (auth)/        # Sign in/up
│       │   └── (main)/        # Feed, messages, jobs, profile, notifications
│       ├── components/        # React components (achievements, messaging, jobs, UI)
│       ├── hooks/             # Custom hooks (auth, feed, socket, notifications)
│       ├── lib/               # API client, WebSocket client, utilities
│       └── types/             # Shared TypeScript types
│
└── shared/                   # Shared constants
```

## Getting Started

### Prerequisites

- **Go** 1.22+
- **Node.js** 22+
- **Docker** and Docker Compose
- **GitHub OAuth App** -- [create one here](https://github.com/settings/developers)
  - Homepage URL: `http://localhost:3000`
  - Callback URL: `http://localhost:3000/auth/callback`
- **Google OAuth Client** (optional, for YouTube) -- [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
  - Authorized redirect URI: `http://localhost:3000/auth/youtube/callback`
  - Enable the YouTube Data API v3
- **Anthropic API key** (optional, for the Smart Feed) -- [console.anthropic.com](https://console.anthropic.com/)

### Setup

1. **Clone and configure environment**

```bash
git clone https://github.com/your-org/praxis.git
cd praxis

# Server environment
cp server/.env.example server/.env
# Edit server/.env with your GitHub OAuth credentials and a JWT secret
```

2. **Start everything with one command**

```bash
make dev
```

This will:
- Start PostgreSQL, Redis, and MinIO via Docker Compose
- Wait for health checks to pass
- Run database migrations automatically
- Launch the Go API server (port 8080), background worker, and Next.js (port 3000) in parallel

3. **Open the app**

Visit [http://localhost:3000](http://localhost:3000) and sign in with GitHub.

### Environment Variables

**Server** (`server/.env`):

```env
SERVER_PORT=8080
DATABASE_URL=postgres://praxis:praxis_dev@localhost:5432/praxis?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-secret-key-here

# GitHub OAuth (required for sign-in)
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Google OAuth — optional, enables YouTube provider
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=

# Optional: GitHub webhook signature secret (POST /api/v1/webhooks/github)
GITHUB_WEBHOOK_SECRET=

# Background scheduler — how often to enqueue provider syncs
SYNC_INTERVAL_HOURS=6

# Anthropic — optional, enables the Smart Feed digest. Without a key the
# /api/v1/feed/smart endpoint returns 503 and the UI shows a graceful
# "not configured" state.
ANTHROPIC_API_KEY=
SMART_FEED_MODEL=claude-haiku-4-5
SMART_FEED_CACHE_TTL=900

FRONTEND_URL=http://localhost:3000
ENVIRONMENT=development
```

**Web** (`web/.env.local`):

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_WS_URL=ws://localhost:8080/ws
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

## Makefile Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start full stack (infra + API + worker + web) |
| `make dev-infra` | Start Docker services only |
| `make dev-server` | Start Go API server only |
| `make dev-worker` | Start background worker only |
| `make dev-web` | Start Next.js dev server only |
| `make test` | Run all tests (server + web) |
| `make test-server` | Run Go tests |
| `make test-web` | Run Vitest unit tests |
| `make test-e2e` | Run Playwright E2E tests |
| `make lint` | Lint Go + TypeScript |
| `make build` | Production build (Go binaries + Next.js) |
| `make migrate-up` | Apply pending database migrations |
| `make migrate-down` | Rollback one migration |
| `make clean` | Remove build artifacts and Docker volumes |

## API Endpoints

### Authentication
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/auth/github` | Initiate GitHub OAuth |
| `POST` | `/api/v1/auth/github/callback` | GitHub OAuth callback |
| `GET` | `/api/v1/auth/youtube` | Initiate Google OAuth (YouTube scope) |
| `POST` | `/api/v1/auth/youtube/callback` | YouTube OAuth callback (links to current user) |
| `POST` | `/api/v1/auth/refresh` | Refresh JWT tokens |
| `POST` | `/api/v1/auth/logout` | Invalidate tokens |
| `GET` | `/api/v1/auth/me` | Current user profile |

The `/auth` group is rate-limited to 20 req/min per IP; everything else is 100 req/min.

### Achievements
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/achievements/sync` | Trigger provider sync |
| `GET` | `/api/v1/achievements/:id` | Get achievement detail |
| `POST` | `/api/v1/achievements/:id/reactions` | Add/toggle reaction |
| `DELETE` | `/api/v1/achievements/:id/reactions` | Remove reaction |

### Feed
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/feed` | Paginated activity feed (self + followed users) |
| `GET` | `/api/v1/feed/smart?refresh=true` | Claude-generated digest with vibe, hero stat, themed chips, milestones, watching, suggested action. Cached per user in Redis (default 15 min). Rate-limited 10 req/min |

### Users
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/users/:username` | User profile |
| `PATCH` | `/api/v1/users/me` | Update own profile |
| `POST` | `/api/v1/users/:username/follow` | Follow user |
| `DELETE` | `/api/v1/users/:username/follow` | Unfollow user |

### Jobs
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/jobs?qualified=true` | List jobs with per-viewer match scores; `qualified=true` filters to fully-qualified |
| `GET` | `/api/v1/jobs/:id` | Job detail with match info for the viewer |
| `POST` | `/api/v1/jobs` | Create job posting |
| `POST` | `/api/v1/jobs/:id/apply` | Apply to job |

Job requirements use the form `TYPE` (any) or `TYPE:N` (threshold/count).
For threshold types like `STARS_MILESTONE:100`, the user qualifies if any
of their achievements has `metadata.current >= 100`. For count types like
`PR_MERGED:5`, they qualify if they have at least 5 of that type.

### Messaging
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/conversations` | List conversations |
| `POST` | `/api/v1/conversations` | Start conversation |
| `GET` | `/api/v1/conversations/search?q=` | Search messages across all the user's conversations |
| `GET` | `/api/v1/conversations/:id/messages` | Get messages |
| `POST` | `/api/v1/conversations/:id/messages` | Send message |
| `DELETE` | `/api/v1/conversations/:id/messages/:messageId` | Delete own message |
| `GET` | `/api/v1/conversations/:id/read-times` | last_read_at per participant (drives read-receipt UI) |
| `PATCH` | `/api/v1/conversations/:id/read` | Mark conversation as read |
| `DELETE` | `/api/v1/conversations/:id` | Delete conversation |
| `WS` | `/ws` | Real-time WebSocket (messages, typing, notifications) |

### Notifications
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/notifications` | List notifications |
| `GET` | `/api/v1/notifications/unread-count` | Unread count |
| `PATCH` | `/api/v1/notifications/:id/read` | Mark as read |

### Webhooks
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/v1/webhooks/github` | GitHub push/repository events. HMAC-SHA256 signature validated against `GITHUB_WEBHOOK_SECRET`. Enqueues a critical-priority `provider:sync` task for the sender |

## Achievement Detection

Praxis automatically detects achievements when a user syncs their connected accounts.

**GitHub provider:**

| Achievement | Trigger | Source ID |
|-------------|---------|-----------|
| **New Repository** | Repo created | `github:repo:{id}` |
| **Stars Milestone** | Repo hits 10/50/100/500/1k stars | `github:stars:{repo}:{n}` |
| **Commit Streak** | 7/30/100 consecutive days with commits | `github:streak:{user}:{n}` |
| **PR Merged** | Pull request merged in external repo | `github:pr:{pr_id}` |
| **First Contribution** | First OSS pull request merged | `github:first-oss:{user}` |
| **Weekly Activity** | Commits in the past 7 days (per-repo breakdown) | `github:weekly-commits:{user}:{week}` |

**YouTube provider:**

| Achievement | Trigger | Source ID |
|-------------|---------|-----------|
| **Channel Created** | Channel exists for the connected Google account | `youtube:channel:{channelId}` |
| **Video Published** | Video on channel | `youtube:video:{videoId}` |
| **Subscribers Milestone** | Channel hits 10/50/100/500/1k/5k/10k/50k/100k/1M subs | `youtube:subscribers:{channelId}:{n}` |
| **Views Milestone** | Video hits 1k/10k/100k/1M/10M views | `youtube:views:{videoId}:{n}` |

Achievements are deduplicated by `source_id` so syncing multiple times won't create duplicates.

When a source becomes unavailable (e.g., a repo is deleted), the achievement is automatically marked as **archived** -- it stays on the profile as proof of past work, but shows "Source no longer available" instead of a proof link.

## Database Schema

13 tables covering the full platform:

- `users`, `auth_accounts`, `connected_providers` -- identity and auth
- `achievements`, `reactions` -- the core achievement system
- `follows` -- social graph
- `conversations`, `conversation_participants`, `messages` -- real-time messaging
- `companies`, `jobs`, `job_applications` -- achievement-based job board
- `notifications` -- activity notifications

## Testing

```bash
# Run everything
make test

# Go unit tests
make test-server

# Frontend unit tests
make test-web

# E2E tests (requires running dev server)
make test-e2e
```

The GitHub provider has comprehensive tests covering repo detection, star milestones, commit streaks, weekly commit tracking (owned + contributed repos), private repo inclusion, and source ID deduplication.

## Roadmap

Detailed phase list lives in [`specs/roadmap.md`](specs/roadmap.md). Current status:

**Shipped (V1, phases 1-15):**
- [x] GitHub OAuth + JWT auth
- [x] GitHub provider (repos, stars, streaks, PRs, weekly commits)
- [x] YouTube provider (channels, videos, subscribers, views)
- [x] Activity feed with infinite scroll, reactions, archiving
- [x] User profiles + follow graph + search
- [x] Real-time messaging (WebSocket, typing indicators, message search)
- [x] Job board with achievement-based match scoring + qualified filter
- [x] Notification system
- [x] Background workers (Asynq queue, periodic scheduler, GitHub webhook)
- [x] Polish & hardening (rate limiting, 404/500 pages, SEO, skeletons)
- [x] CI (GitHub Actions: Postgres + Redis services, race tests, build)
- [x] Smart Feed — Claude-generated weekly digest with hero stat, vibe, themed chips, milestones, watching, suggested action, and linked achievements

**Up next (V2):**
- [ ] Phase 16 — Cryptographic verification (Ed25519 signatures, public verify endpoint)
- [ ] Phase 17 — Responsive + accessibility audit
- [ ] Phase 18 — Media uploads (MinIO/S3)
- [ ] Phase 19 — Transactional email (Resend / SES)
- [ ] Phase 20 — Playwright e2e suite
- [ ] Phase 21+ — Company pages, hiring workflow, analytics, observability, prod deploy

## License

MIT
