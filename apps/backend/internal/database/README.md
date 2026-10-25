# Database Package: Connection and Schema Lifecycle

Back to [../../docs/README.md](../../docs/README.md). For SQL fundamentals, begin with [../../../../docs/database/README.md](../../../../docs/database/README.md).

## Two Different Responsibilities

| File | What it does | When it runs |
| --- | --- | --- |
| [database.go](database.go) | Builds a DSN, configures tracing, creates a pgx pool, pings PostgreSQL, and exposes `Close` | Normal server construction |
| [migrator.go](migrator.go) | Opens a single pgx connection and uses Tern to apply embedded SQL files | Non-local startup and explicitly invoked test setup |
| [migrations/001_setup.sql](migrations/001_setup.sql) | Shared SQL functions used by later queries/triggers | Migration 1 |
| [migrations/002_todos.sql](migrations/002_todos.sql) | Tasker tables, indexes, triggers, and hierarchy constraint | Migration 2 |

Repositories are consumers of the pool; they contain feature queries, not connection setup. Models describe Go data; they do not auto-migrate these tables.

## Read database.New

1. Join host and port and assemble a PostgreSQL connection string from config.
2. Parse it using `pgxpool.ParseConfig`.
3. Attach optional New Relic query instrumentation and, in local mode, a Zerolog trace adapter.
4. Create the pool and wrap it in `Database`.
5. Ping with a ten-second timeout, returning an error if the database is unreachable.

`multiTracer` forwards query-start/query-end events to configured tracers. Its type assertions check whether each stored value implements the relevant hook. This is instrumentation around a query, not another SQL execution.

The config fields for 25 connections and 300-second lifetimes are **not applied** by this constructor. The pool currently uses pgx defaults unless settings reach it through its parsed configuration. Read the constructor rather than inferring active behavior from an environment variable name.

The current DSN uses manual formatting and `url.QueryEscape` for the password. That is not a general guarantee of correct URL-userinfo encoding; credentials containing spaces or other special characters need tests and a proper structured URL construction approach when changing this code. Never log real DSNs.

## Read database.Migrate

`//go:embed migrations/*.sql` includes the SQL in the compiled binary. `fs.Sub` exposes the migration directory to Tern; `NewMigrator(..., "schema_version")` uses a database version table; `Migrate` applies pending migrations in order.

This uses one connection because it is a serial schema-management operation, not the request-serving pool. The SQL is embedded at build time, so an old binary will not gain migrations added after it was built.

Locally, main skips auto-migration. The Taskfile invokes the separate Tern CLI and needs its own DSN input. `go mod download` fetches a library dependency; it does not install the CLI executable on PATH.

## What to Read Next

- [../../../../docs/database/03_migrations_with_tern.md](../../../../docs/database/03_migrations_with_tern.md): ordering, up/down sections, and commands.
- [../repository/README.md](../repository/README.md): named parameters, row mapping, and the actual feature queries.
- [../testing/README.md](../testing/README.md): applying migrations to a disposable database.

Creating a migration file, applying it successfully, and exposing a working HTTP feature are three separate milestones.