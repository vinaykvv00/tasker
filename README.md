# Tasker

A full-stack task manager built while learning from a tutorial and extending a
Go/TypeScript boilerplate. The Go API, React UI and shared packages are present;
testing, unfinished UI behavior and deployment hardening remain ongoing work.

## Understand the Project

Start with the **[complete learning path](docs/overall_end_to_end/README.md)**.
It explains the current source, not an idealized finished system.

| Guide | What you will learn |
| --- | --- |
| [Architecture](docs/overall_end_to_end/architecture_and_tech_stack.md) | Processes, technologies, layers and design tradeoffs |
| [Request walkthroughs](docs/overall_end_to_end/request_walkthroughs.md) | Browser-to-SQL flow, uploads, updates and errors |
| [Backend](apps/backend/docs/README.md) | Go syntax, startup, each layer and background jobs |
| [Database](docs/database/README.md) | PostgreSQL, pgx, schema, migrations and local inspection |
| [Frontend](docs/frontend/README.md) | Routes, providers, forms, query cache and current placeholders |
| [Shared packages](packages/README.md) | Zod, contracts, generated API docs and email templates |
| [Learning workbook](docs/overall_end_to_end/learning_workbook.md) | Exercises, glossary and a method for building independently |

## Features

- Todo CRUD, filtering, pagination, status, priority, categories and comments.
- Attachment metadata in PostgreSQL and file bytes in S3-compatible storage.
- Clerk authentication in React and authenticated Go feature routes.
- React Query data fetching, forms, dashboard and theme switching.
- Tern migrations, pgx pooling, Asynq email tasks and separate cron commands.
- Structured logs and optional New Relic integration.
- Shared TypeScript contracts, generated OpenAPI docs and React Email sources.

Settings import/export/delete-all, notification preferences and bulk mutations
are not complete. No backend test files are present yet; Testcontainers helpers
are groundwork, not coverage. See the learning guides for known limitations.

## Project Structure

```text
apps/backend/       Go API, cron executable, migrations and email HTML
apps/frontend/      React single-page application
packages/zod/       Shared TypeScript data schemas
packages/openapi/   ts-rest contracts and OpenAPI generator
packages/emails/    React Email template sources
docs/               Architecture, database and learning guides
docker-compose.yml  Local PostgreSQL and Redis
```

## Quick Start

### Prerequisites

- Go 1.24.5 or newer
- Node.js 22+ and Bun
- Docker/Compose, or compatible PostgreSQL and Redis services
- Task and the Tern CLI for the documented backend workflow
- Clerk configuration and the external integration credentials required by Go config

Read [local setup and its security notes](apps/backend/docs/local-development.md)
before starting services. Compose currently publishes passwordless database/Redis
ports on all interfaces; it is not a hardened deployment configuration.

### Setup

Clone only when you need another working copy:

```bash
git clone https://github.com/vinaykvv00/tasker.git
cd tasker
bun install
```

Create an ignored backend environment file from
[the sample](apps/backend/.env.sample) **only if it does not already exist**.
Use uppercase dotted names such as `TASKER_DATABASE.HOST`; local Redis needs
`localhost:6379`, not a `redis://` URL. Do not source dotted assignments as Bash.

Configure the frontend's ignored local environment with
`VITE_CLERK_PUBLISHABLE_KEY`, `VITE_API_URL="http://localhost:8080"`, and
`VITE_ENV="local"`. Browser values are public; never put backend secrets there.

After reviewing configuration, start data services from the root:

```bash
docker compose up -d
cd apps/backend
go mod download
```

Apply reviewed migrations to the intended **local** database, then start the API:

```bash
task migrations:up TASKER_DB_DSN="postgres://postgres@localhost:5432/tasker?sslmode=disable"
task run
```

Migration execution changes the database after confirmation. Local API startup
does not apply migrations automatically.

In another terminal at the root:

```bash
bun run dev
```

This starts JS workspace development processes, not the Go API. The frontend is
normally `http://localhost:3000`, API health is `http://localhost:8080/status`,
and API documentation is `http://localhost:8080/docs`.

## Development

| Directory | Command | Purpose |
| --- | --- | --- |
| Backend | `task --list-all` | Available tasks |
| Backend | `task migrations:new name=example` | Create, not apply, a migration |
| Backend | `go build ./...` / `go vet ./...` | Compile and static checks |
| Backend | `go test ./...` | Run tests once added; no tests is not coverage |
| Backend | `go run ./cmd/cron list` | Inspect named periodic jobs without executing them |
| Root | `bun run build --filter=@tasker/openapi` | Build shared contracts and dependencies |
| Root | `bun run --cwd packages/openapi gen` | Regenerate API JSON artifacts |
| Root | `bun run --cwd packages/emails export` | Regenerate backend email HTML |
| Frontend | `bunx tsc -b --pretty false` | Check referenced TypeScript projects |
| Frontend | `bun run lint` / `bun run build` | Lint / bundle the UI |

The frontend's current plain `tsc` script does not check its project references.
Use build mode for a real typecheck. Root build does not run Go, API generation
or email export. Check scripts before assuming what a command covers.

## Before Deployment

Finish behavior-focused tests, review ownership and upload limits, configure
TLS and credentials, restrict network bindings, verify migrations and artifact
packaging, establish backups/restore, and deliberately schedule cron. Queue
retries and logging do not replace those checks. Deployment is not configured
or validated by this documentation update.

## Contributing

Use the [feature checklist](docs/overall_end_to_end/how_to_add_features.md).
Keep changes focused, preserve existing data/migration history, and use short
lowercase commit messages grouped by behavior. Do not commit environment secrets
or the ignored tutorial-reference checkout.

## License

See [LICENSE](LICENSE) for the repository's license terms.
