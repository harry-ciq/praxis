# Praxis

**Where execution speaks louder than words.**

Praxis is a professional network driven by verified execution, not self-promotion. Instead of users writing about what they did, Praxis connects to their GitHub, YouTube, and other platforms to automatically detect, verify, and surface real achievements -- repos created, PRs merged, commit streaks, stars milestones, and more.

Think LinkedIn, but every achievement on your profile is backed by proof.

## Why Praxis?

The problem with existing professional networks: anyone can claim anything. Praxis flips this by making the platform itself the verifier. Your profile is built from what you actually ship, not what you say you shipped.

- **Automatic achievement detection** -- connect GitHub and Praxis tracks your repos, commits, PRs, stars, and streaks
- **Verified proof** -- every achievement links back to its source with cryptographic verification (Phase 2)
- **Achievement-based job matching** -- employers post requirements like "maintained a 30-day commit streak" or "has a repo with 100+ stars", and Praxis matches verified candidates
- **Real-time feed** -- see what builders in your network are actually shipping

## Tech Stack

| Layer | Technology |
|-------|-----------|
| **Frontend** | Next.js 16 (App Router), React 19, TypeScript, Tailwind CSS v4, shadcn/ui |
| **Backend** | Go 1.22+, Chi router, JWT auth |
| **Database** | PostgreSQL 16 (pgx driver), Redis 7 |
| **Real-time** | WebSocket (gorilla/websocket), Redis pub/sub |
| **Background Jobs** | Asynq (Redis-backed task queue) |
| **External APIs** | GitHub API (go-github), YouTube API (planned) |
| **Object Storage** | MinIO (S3-compatible) |
| **Testing** | Go testify, Vitest + React Testing Library, Playwright |

## Architecture

```
                    +-------------------------+
                    |    Next.js Frontend     |
                    |    Port 3000            |
                    +-----------+-------------+
                                |
                       REST + WebSocket
                                |
                    +-----------v-------------+
                    |     Go API Server       |
                    |     Port 8080           |
                    |                         |
                    |  Handlers -> Services   |
                    |  -> Repositories (pgx)  |
                    |                         |
                    |  Provider Registry      |
                    |   +- GitHubProvider     |
                    |   +- YouTubeProvider    |
                    +-+--------+----------+---+
                      |        |          |
               +------v-+  +--v---+  +---v----+
               |Postgres |  |Redis |  | MinIO  |
               +----+----+  +--+---+  +--------+
                    |           |
                    |     +-----v--------+
                    |     | Asynq Worker |
                    |     | - SyncGitHub |
                    |     | - SendEmail  |
                    |     +--------------+
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
│   │   └── worker/main.go    # Background job worker
│   ├── internal/
│   │   ├── config/            # Environment config with validation
│   │   ├── handler/           # HTTP handlers (auth, feed, achievements, jobs, messages)
│   │   ├── service/           # Business logic layer
│   │   ├── repository/        # Database access (pgx)
│   │   ├── provider/          # Achievement providers (GitHub, YouTube)
│   │   ├── middleware/        # Auth, CORS, logging, rate limiting
│   │   ├── ws/                # WebSocket hub for real-time messaging
│   │   └── worker/            # Async job processors
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
REDIS_URL=localhost:6379
JWT_SECRET=your-secret-key-here
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
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
| `POST` | `/api/v1/auth/github/callback` | OAuth callback |
| `POST` | `/api/v1/auth/refresh` | Refresh JWT tokens |
| `POST` | `/api/v1/auth/logout` | Invalidate tokens |
| `GET` | `/api/v1/auth/me` | Current user profile |

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
| `GET` | `/api/v1/feed` | Paginated activity feed |

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
| `GET` | `/api/v1/jobs` | List jobs |
| `GET` | `/api/v1/jobs/:id` | Job detail |
| `POST` | `/api/v1/jobs` | Create job posting |
| `POST` | `/api/v1/jobs/:id/apply` | Apply to job |

### Messaging
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/conversations` | List conversations |
| `POST` | `/api/v1/conversations` | Start conversation |
| `GET` | `/api/v1/conversations/:id/messages` | Get messages |
| `POST` | `/api/v1/conversations/:id/messages` | Send message |
| `WS` | `/ws` | Real-time WebSocket |

### Notifications
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/v1/notifications` | List notifications |
| `GET` | `/api/v1/notifications/unread-count` | Unread count |
| `PATCH` | `/api/v1/notifications/:id/read` | Mark as read |

## Achievement Detection

Praxis automatically detects achievements when a user syncs their connected accounts.

| Achievement | Trigger | Source ID |
|-------------|---------|-----------|
| **New Repository** | Repo created | `github:repo:{id}` |
| **Stars Milestone** | Repo hits 10/50/100/500/1k stars | `github:stars:{repo}:{n}` |
| **Commit Streak** | 7/30/100 consecutive days with commits | `github:streak:{user}:{n}` |
| **PR Merged** | Pull request merged in external repo | `github:pr:{pr_id}` |
| **First Contribution** | First OSS pull request merged | `github:first-oss:{user}` |
| **Weekly Activity** | Commits in the past 7 days (per-repo breakdown) | `github:weekly-commits:{user}:{week}` |

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

- [x] GitHub OAuth + JWT auth
- [x] GitHub achievement provider (repos, stars, streaks, PRs, weekly commits)
- [x] Activity feed with infinite scroll
- [x] Reactions (clap, fire, rocket) with optimistic updates
- [x] Achievement archiving when source is deleted
- [x] Real-time messaging (WebSocket)
- [x] Job board with achievement-based matching
- [x] Notification system
- [ ] YouTube provider (videos, subscribers, views milestones)
- [ ] Blockchain verification (Phase 2)
- [ ] Webhook-based real-time achievement detection
- [ ] Open Graph image generation for shared achievements
- [ ] Mobile-responsive polish

## License

MIT
