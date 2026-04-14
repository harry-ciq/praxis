.PHONY: dev dev-infra dev-server dev-web test test-server test-web lint build clean migrate-up migrate-down sqlc

# Start all infrastructure services
dev-infra:
	docker compose up -d

# Start Go API server
dev-server:
	cd server && go run ./cmd/api

# Start Go worker
dev-worker:
	cd server && go run ./cmd/worker

# Start Next.js dev server
dev-web:
	cd web && npm run dev

# Start everything in one command (all services run in background, logs merged)
dev: dev-infra
	@echo "⏳ Waiting for infrastructure..."
	@until docker compose exec -T postgres pg_isready -U praxis > /dev/null 2>&1; do sleep 1; done
	@until docker compose exec -T redis redis-cli ping > /dev/null 2>&1; do sleep 1; done
	@echo "✅ PostgreSQL + Redis + MinIO ready"
	@$(HOME)/go/bin/migrate -path server/migrations -database "postgres://praxis:praxis_dev@localhost:5432/praxis?sslmode=disable" up 2>/dev/null || true
	@echo "✅ Migrations applied"
	@echo "🚀 Starting Go API server (port 8080), Go worker, and Next.js (port 3000)..."
	@trap 'kill 0' INT TERM EXIT; \
		(cd server && go run ./cmd/api 2>&1 | sed 's/^/[api]    /') & \
		(cd server && go run ./cmd/worker 2>&1 | sed 's/^/[worker] /') & \
		(cd web && npm run dev 2>&1 | sed 's/^/[web]    /') & \
		wait

# Run all tests
test: test-server test-web

test-server:
	cd server && go test ./... -v -count=1

test-web:
	cd web && npm run test

test-e2e:
	cd web && npm run test:e2e

# Lint
lint:
	cd server && golangci-lint run ./...
	cd web && npm run lint

# Build
build:
	cd server && go build -o bin/api ./cmd/api && go build -o bin/worker ./cmd/worker
	cd web && npm run build

# Database migrations
migrate-up:
	cd server && go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate \
		-path migrations -database "postgres://praxis:praxis_dev@localhost:5432/praxis?sslmode=disable" up

migrate-down:
	cd server && go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate \
		-path migrations -database "postgres://praxis:praxis_dev@localhost:5432/praxis?sslmode=disable" down 1

# Generate sqlc code
sqlc:
	cd server && sqlc generate

# Clean
clean:
	rm -rf server/bin web/.next web/node_modules
	docker compose down -v
