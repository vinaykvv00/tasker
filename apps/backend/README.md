# Tasker Go Backend

A Go/Echo backend built from a boilerplate and extended with Todo, Category, and Comment models and repositories. Feature services and HTTP routes are still being built. Existing helpers are not a production-readiness guarantee.

## Start Learning Here

**[docs/README.md](docs/README.md)** is the beginner reading path, folder/file map, and current feature-status overview.

| Read | For |
| --- | --- |
| [docs/go-reading-basics.md](docs/go-reading-basics.md) | Go syntax as used in this code |
| [internal/server/README.md](internal/server/README.md) | Startup, shared resources, and shutdown |
| [internal/model/README.md](internal/model/README.md) | Models, DTOs, tags, pointers, and defaults |
| [internal/repository/README.md](internal/repository/README.md) | SQL, pgx, mapping, filtering, and CRUD |
| [internal/service/README.md](internal/service/README.md) | Current wiring and planned business rules |
| [docs/building-a-feature.md](docs/building-a-feature.md) | Connect an operation to authenticated HTTP |
| [docs/local-development.md](docs/local-development.md) | Run, inspect, and troubleshoot safely |

Each major internal folder has a README mapping its source files. Shared patterns are explained once rather than repeated for every CRUD method.

## Architecture Overview

The intended request flow is router/middleware -> handler -> service -> repository -> pgxpool -> PostgreSQL. Models are values passed between these layers, not another running service.

```
backend/
├── cmd/tasker/           # Application entry point
├── internal/                  # Private application code
│   ├── config/               # Configuration management
│   ├── database/             # Database connections and migrations
│   ├── handler/              # HTTP request handlers
│   ├── service/              # Business logic layer
│   ├── repository/           # Data access layer
│   ├── model/                # Domain models
│   ├── middleware/           # HTTP middleware
│   ├── lib/                  # Shared libraries
│   └── validation/           # Request validation
├── static/                   # Static files (OpenAPI spec)
├── templates/                # Email templates
└── Taskfile.yml              # Task automation
```

## Current State

The executable is [cmd/tasker/main.go](cmd/tasker/main.go). It constructs config, logging, a PostgreSQL pool, a Redis client, an Asynq worker, registries, and Echo. The worker currently runs inside the Go process; PostgreSQL and Redis run externally.

The service registry exposes Auth and Job. The HTTP registry exposes Health and OpenAPI. Todo, Category, and Comment repositories are constructed, but are not connected through feature services and routes. Clerk middleware exists but is not attached to feature routes yet.

The current HTTP surface is `GET /status`, `GET /docs`, and `/static/*`. There is no backend root page or registered Todo API. Inspect individual health checks: Redis failure currently does not change the overall healthy/200 response.

## Getting Started

Follow [docs/local-development.md](docs/local-development.md), including its security notes, before starting infrastructure. Current Compose uses passwordless services and broadly published ports; it is not a deployment template.

Use Go 1.24.5 or newer, Task, Docker/Compose, and the Tern CLI for manual migrations. Run backend commands from this directory. Keep credentials in an ignored local environment file; [.env.sample](.env.sample) is a template, not a guarantee of working values.

Configuration keys use the `tasker_` prefix and dots, such as `tasker_DATABASE.HOST`. Main skips automatic migrations in local mode. The Taskfile migration command needs its own DSN.

## Development

### Available Tasks

| Command | Effect |
| --- | --- |
| `task --list-all` | List actual tasks |
| `task run` | Run `go run ./cmd/tasker` |
| `task migrations:new name=example` | Create a migration file, without applying it |
| `task migrations:up tasker_DB_DSN="..."` | Apply reviewed migrations to the selected database after confirmation |
| `go test ./...` | Run available Go tests; tests using containers require Docker |
| `task tidy` | Format Go, tidy dependencies, and verify modules; changes files |

There is no current `task test` or `migrations:down` task.

## Before Extending or Deploying

Read the current limitations beside each folder's explanation, particularly DTO validation, user-scoped search SQL, populated Todo joins, handler request allocation, and unused pool settings. Add tests for those behaviors rather than relying on startup logs.

For general database concepts, see [../../docs/database/README.md](../../docs/database/README.md). For safe test isolation, see [internal/testing/README.md](internal/testing/README.md). Logging and telemetry are explained in [internal/logger/README.md](internal/logger/README.md).

The frontend loading, a successful DB ping, or a helper existing on disk does not verify the entire application's behavior. Keep code, migrations, contracts, and tests aligned as you add each feature.

## License

See [../../LICENSE](../../LICENSE).