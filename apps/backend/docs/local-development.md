# Local Development and Troubleshooting

Back to [README.md](README.md). These are instructions, not commands executed as part of writing this guide.

## What Runs Where?

| Component | How it runs in this setup |
| --- | --- |
| Go backend | `task run` in the WSL backend directory |
| Frontend/package watchers | `bun dev` from the repository root |
| PostgreSQL and Redis | Containers described by [../../../docker-compose.yml](../../../docker-compose.yml) |
| Docker CLI in WSL | Connects to the selected Docker engine; with Desktop integration this can be Docker Desktop's Linux engine |

The CLI and the engine are different. Check `docker context show` and `docker info` rather than assuming which engine owns a container. Compose describes the services; the Docker engine runs them. Data lives in Docker volumes, not in the Go repository.

## Security Before Startup

Current Compose uses PostgreSQL `trust` authentication, Redis without an application password, and mappings `5432:5432` / `6379:6379` on all host interfaces. This is **not** a hardened local configuration. Before use on a shared/untrusted network, restrict published bindings to loopback and configure authentication deliberately. The Redis image is a milestone/prerelease tag and should be reviewed before deployment. These docs do not change or restart existing containers.

Named volumes preserve data across container recreation; they are not backups. Do not run `docker compose down -v` or delete volumes as routine troubleshooting: that can delete your database and queued data.

## Prerequisites and Startup

The Go module requires Go 1.24.5 or newer. Use Bun for the workspace, Docker with Compose for infrastructure, Task for the backend commands, and the Tern CLI for manual migrations.

After reviewing the configuration, from the repository root:

```bash
docker compose up -d
docker compose ps
cd apps/backend
go mod download
```

Create a local environment file from [../.env.sample](../.env.sample) only if you do not already have one; never overwrite working credentials. Match the dotted key names described in [../internal/config/README.md](../internal/config/README.md).

Relevant local connection values are PostgreSQL host `localhost`, port `5432`, database `tasker`, user `postgres`, and Redis address `localhost:6379`. The current Redis clients expect no `redis://` prefix, even if a sample file still uses it.

To install the matching Tern CLI and apply reviewed migrations from the backend directory:

```bash
go install github.com/jackc/tern/v2@v2.3.3
export PATH="$(go env GOPATH)/bin:$PATH"
tern version
task migrations:up TASKER_DB_DSN="postgres://postgres@localhost:5432/tasker?sslmode=disable"
task run
```

If `GOBIN` is customized, add that directory to PATH instead. The migration command changes the database after confirmation. Main does not apply migrations automatically when `Primary.Env` is `local`.

In a separate terminal at the repository root, use `bun install` if dependencies are missing, then `bun dev`. That workspace command does not launch the Go API.

## Observe Without Modifying Data

```bash
curl -i http://localhost:8080/status
docker compose exec postgres pg_isready -U postgres -d tasker
docker compose exec redis redis-cli ping
docker compose exec postgres psql -U postgres -d tasker -c '\dt'
docker compose exec redis redis-cli --scan
```

Container checks prove the service inside a container responds, not that every application query works. Read individual `/status` checks: Redis failure currently does not change the overall healthy/200 result.

The schema listing should show feature tables only after migrations. `redis-cli --scan` incrementally lists keys; do not use destructive flush commands to inspect a queue.

## GUI Connections

VS Code database tools in WSL, or Windows pgAdmin/TablePlus with working port forwarding, can normally use `127.0.0.1:5432`, database `tasker`, user `postgres`. The current trust setup accepts no password; update the client if you later configure authentication.

A Redis-capable client can connect to `127.0.0.1:6379`. Windows-to-WSL/Desktop forwarding depends on the actual setup and firewall. If localhost fails, verify the selected engine, port mapping, and service state before changing network exposure.

## Symptom to First Check

| Symptom | Check first |
| --- | --- |
| All config fields missing | Backend working directory, dotenv syntax, exact prefix/dotted names |
| `too many colons in address` for Redis | Remove URL scheme for the current host/port consumer |
| `Task "name:todos" does not exist` | Use `task migrations:new name=todos` |
| `tern` not found | Install binary and put Go's install directory on PATH |
| `no sql in forward migration step` | Executable SQL must precede the Tern separator |
| Relation/function not found | Target database and applied migration version |
| 404 at `/` | No backend root page is registered; try `/status` or `/docs` |
| 404 at a Todo URL | Check method and `/api/v1` prefix against registered routes |
| SQL or validation fails but Go builds | SQL strings and tags require runtime tests |
| Frontend runs but requests fail | Browser Network tab: actual URL, prefix, CORS, token, method, response |

Run `task --list-all` for tasks that actually exist. The current Taskfile has no `test` or `migrations:down` task; use Go test commands or deliberately reviewed Tern commands instead.

## Frontend Configuration and Checks

The frontend dev script loads its ignored local environment file via `env-cmd`.
It needs `VITE_CLERK_PUBLISHABLE_KEY`, `VITE_API_URL` (the backend origin, normally
`http://localhost:8080`, without `/api`), and optionally `VITE_ENV`.
Use the matching Clerk application on both sides. Never put a Clerk secret key,
Resend key or storage credentials into browser variables.

From the repository root, build the shared contracts before standalone frontend
work:

```bash
bun install
bun run build --filter=@tasker/openapi
cd apps/frontend
bunx tsc -b --pretty false
bun run lint
bun run dev
```

The frontend's plain `tsc` script has a solution config with `files: []` and
project references. It does not check the referenced app by itself. `tsc -b`
checks those projects. A Vite build transpiles and bundles; it is not evidence
of a successful full TypeScript check.

For the backend, `go build ./...` checks compilation and `go vet ./...` performs
static analysis. Neither executes SQL queries or proves authentication works.
Testing and deployment remain separate learning stages.