# Chapter 2: PostgreSQL and the pgx Driver

Chapter 1 covered *what* a database is. This chapter covers **how your Go code actually talks to it** — the driver, the connection pool, and the DSN.

---

## 2.1 PostgreSQL is a Server

PostgreSQL is a **long-running server process** that listens on a network port (default `5432`). It is not a library you link into your app.

```mermaid
graph LR
    A["Go App<br/>(client)"] -->|"TCP :5432"| B["PostgreSQL<br/>(server)"]
    B --> C[("Data files<br/>on disk")]
```

Because it's a server, it can be:

- On the same machine (`localhost`)
- In a Docker container (this project)
- On a remote cloud host (production)

Your app doesn't care — it just needs an **address** and **credentials**.

### In this project

PostgreSQL runs in a Docker container defined in [docker-compose.yml](../../docker-compose.yml):

```yaml
postgres:
  image: postgres:16-alpine
  container_name: tasker-postgres
  environment:
    - POSTGRES_DB=tasker
    - POSTGRES_USER=postgres
    - POSTGRES_HOST_AUTH_METHOD=trust
  ports:
    - "5432:5432"
```

The `ports: "5432:5432"` line maps the container's port to your host, so `localhost:5432` reaches it.

Security caveat: that mapping is not restricted to loopback, and `trust` disables password authentication for matching connections. This is a local learning configuration, not a safe network-facing deployment. Restrict exposure and use authentication before sharing it.

---

## 2.2 Why a Driver is Needed

Sending plain SQL text to a TCP socket is insufficient: PostgreSQL expects a handshake and protocol messages. A **driver** implements this protocol and value conversion so you do not have to write them yourself.

The driver's responsibilities:

| Job | Description |
| :--- | :--- |
| **Connect** | Open a TCP connection and perform the startup handshake |
| **Encode** | Convert Go values → Postgres wire format |
| **Decode** | Convert Postgres results → Go values |
| **Pool** | Reuse connections efficiently |
| **Trace** | Report query timings to observability tools |

### Why `pgx` and not `database/sql`?

Go's standard `database/sql` provides a generic database API and its own connection pool through `sql.DB`; it still needs a driver. pgx provides both a native PostgreSQL API and an adapter for `database/sql`.

This project uses the native `pgx/v5` API with `pgxpool`, giving direct access to PostgreSQL types, named arguments, row collectors, batching, and tracing. Neither API is universally faster; measure actual workloads rather than assuming a library choice guarantees performance.

---

## 2.3 The DSN (Data Source Name)

The **DSN** is a single string containing everything needed to connect:

```text
postgres://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=MODE
postgres://postgres:@localhost:5432/tasker?sslmode=disable
```

| Part | Meaning | Source in this project |
| :--- | :--- | :--- |
| `postgres://` | PostgreSQL connection-URI scheme | hardcoded |
| `postgres` | Username | `tasker_DATABASE.USER` |
| *(empty)* | Password | `tasker_DATABASE.PASSWORD` |
| `localhost` | Host | `tasker_DATABASE.HOST` |
| `5432` | Port | `tasker_DATABASE.PORT` |
| `tasker` | Database name | `tasker_DATABASE.NAME` |
| `sslmode=disable` | No TLS (local only) | `tasker_DATABASE.SSL_MODE` |

### How this project builds the DSN

From [database.go](../../apps/backend/internal/database/database.go):

```go
hostPort := net.JoinHostPort(cfg.Database.Host, strconv.Itoa(cfg.Database.Port))

// URL-encode the password
encodedPassword := url.QueryEscape(cfg.Database.Password)
dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s",
    cfg.Database.User,
    encodedPassword,
    hostPort,
    cfg.Database.Name,
    cfg.Database.SSLMode,
)
```

Two details worth noting:

1. **`net.JoinHostPort`** correctly handles IPv6 addresses (adds brackets when needed).
2. **`url.QueryEscape`** is the current implementation, but query escaping is not the same as URL-userinfo encoding (spaces are one example). This manual DSN construction needs care for arbitrary credentials; structured URL userinfo is preferable when revising it.

> **Security note:** use an appropriately authenticated TLS connection for remote databases, normally `verify-full` with a trusted CA. `require` encrypts traffic but is not a general substitute for server identity verification. Do not print real DSNs in logs.

---

## 2.4 Connection Pooling

Opening a database connection is **expensive**: TCP handshake, TLS negotiation, authentication, and server-side setup. Doing this per request would be slow.

A **connection pool** keeps a set of connections open and reuses them.

```mermaid
graph TD
    A["Incoming requests"] --> P["pgxpool"]
    P --> C1["Connection 1"]
    P --> C2["Connection 2"]
    P --> C3["Connection 3"]
    P --> C4["... up to pgxpool MaxConns"]
    C1 --> DB[("PostgreSQL")]
    C2 --> DB
    C3 --> DB
    C4 --> DB
```

When a request needs the database:

1. It **borrows** a connection from the pool.
2. It runs its query.
3. It **returns** the connection to the pool.

If all connections are busy, the request **waits** until one frees up.

### Declared pool settings vs. actual behavior

From `.env`:

```env
tasker_DATABASE.MAX_OPEN_CONNS="25"
tasker_DATABASE.MAX_IDLE_CONNS="25"
tasker_DATABASE.CONN_MAX_LIFETIME="300"
tasker_DATABASE.CONN_MAX_IDLE_TIME="300"
```

These values exist in configuration, but [database.go](../../apps/backend/internal/database/database.go) does not assign them to the pgx pool config. They are not the active connection limits or timeouts.

| Setting | Intended meaning, not currently applied |
| :--- | :--- |
| `MAX_OPEN_CONNS` | Maximum simultaneous connections (25) |
| `MAX_IDLE_CONNS` | How many idle connections to keep warm (25) |
| `CONN_MAX_LIFETIME` | Recycle a connection after 300s (avoids stale connections) |
| `CONN_MAX_IDLE_TIME` | Close idle connections after 300s |

pgxpool uses fields such as `MaxConns`, `MinConns`, `MaxConnLifetime`, and `MaxConnIdleTime`; it does not have the exact same max-idle-count API as `database/sql`. The current constructor relies on pgx defaults. A pool also does not necessarily pre-open its maximum number of connections.

### Creating the pool

```go
pgxPoolConfig, err := pgxpool.ParseConfig(dsn)
if err != nil {
    return nil, fmt.Errorf("failed to parse pgx pool config: %w", err)
}

pool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)
if err != nil {
    return nil, fmt.Errorf("failed to create pgx pool: %w", err)
}
```

### Verifying the connection

After creating the pool, the code **pings** the database to fail fast if something is wrong:

```go
ctx, cancel := context.WithTimeout(context.Background(), DatabasePingTimeout*time.Second)
defer cancel()
if err = pool.Ping(ctx); err != nil {
    return nil, fmt.Errorf("failed to ping database: %w", err)
}

logger.Info().Msg("connected to the database")
```

This is why you see `INF connected to the database` in your server logs.

---

## 2.5 Running Queries

Once you have a pool, you run queries through it. The following are illustrative fragments, not complete functions or implemented HTTP endpoints. Resource queries should include ownership predicates.

### `QueryRow` - read the first row, or report no rows

```go
var title string
err := db.Pool.QueryRow(ctx,
    "SELECT title FROM todos WHERE id = $1 AND user_id = $2", id, userID,
).Scan(&title)
```

### `Query` — expect many rows

```go
rows, err := db.Pool.Query(ctx,
    "SELECT id, title FROM todos WHERE user_id = $1", userID)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var id uuid.UUID
    var title string
    if err := rows.Scan(&id, &title); err != nil {
        return err
    }
    // use id, title
}
return rows.Err() // always check after the loop
```

### `Exec` — no rows returned

```go
tag, err := db.Pool.Exec(ctx,
    "UPDATE todos SET status = $1 WHERE id = $2 AND user_id = $3", "completed", id, userID)
// tag.RowsAffected() tells you how many rows changed
```

### Parameter placeholders

Notice `$1`, `$2` — these are **positional placeholders**. The driver sends the SQL and the values **separately**, so the database never confuses data for code. This is the correct, injection-safe way to pass values.

```go
// ✅ SAFE
db.Pool.QueryRow(ctx, "SELECT * FROM todos WHERE id = $1", id)

// ❌ NEVER DO THIS — SQL injection
db.Pool.QueryRow(ctx, "SELECT * FROM todos WHERE id = '" + id + "'")
```

---

## 2.6 Transactions

For multi-step operations that must be atomic, use a transaction:

```go
tx, err := db.Pool.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx) // no-op if already committed

_, err = tx.Exec(ctx, "UPDATE todos SET status = $1 WHERE id = $2 AND user_id = $3", "completed", id, userID)
if err != nil {
    return err // Rollback runs via defer
}

_, err = tx.Exec(ctx, "INSERT INTO todo_comments (todo_id, user_id, content) VALUES ($1, $2, $3)",
    id, userID, "Completed!")
if err != nil {
    return err
}

return tx.Commit(ctx)
```

The deferred rollback cleans up an uncommitted transaction on ordinary early returns. This fragment illustrates transaction mechanics; a real use case must also check ownership/existence, affected rows, and completion-timestamp rules. Calls through a separate pool rather than `tx` are not part of this transaction.

---

## 2.7 Tracing and Observability

This project attaches **tracers** to the pool so every query is measured.

From `database.go`:

```go
// Add New Relic PostgreSQL instrumentation
if loggerService != nil && loggerService.GetApplication() != nil {
    pgxPoolConfig.ConnConfig.Tracer = nrpgx5.NewTracer()
}
```

In local development, a **second** tracer logs queries to the console:

```go
if cfg.Primary.Env == "local" {
    globalLevel := logger.GetLevel()
    pgxLogger := loggerConfig.NewPgxLogger(globalLevel)
    // Chain tracers - New Relic first, then local logging
    ...
}
```

Because `pgx` allows only one tracer, the project defines a `multiTracer` that fans out to several:

```go
type multiTracer struct {
    tracers []any
}
```

This is why you can see SQL timings in your terminal **and** in New Relic simultaneously.

---

## 2.8 The `Database` Struct

Everything is wrapped in a small struct:

```go
type Database struct {
    Pool *pgxpool.Pool
    log  *zerolog.Logger
}
```

- `Pool` — the connection pool, used by repositories
- `log` — a logger for lifecycle messages

And a `Close` method for graceful shutdown:

```go
func (db *Database) Close() error {
    db.log.Info().Msg("closing database connection pool")
    db.Pool.Close()
    return nil
}
```

---

## 2.9 Summary

| Concept | Meaning | In this project |
| :--- | :--- | :--- |
| PostgreSQL server | Separate process listening on a port | Docker container, port 5432 |
| Driver | Library speaking the wire protocol | `pgx/v5` |
| DSN | Connection string | Built from `.env` in `database.go` |
| Pool | Reusable set of connections | `pgxpool`; declared 25-connection config is currently unused |
| Placeholder | Safe value substitution | `$1`, `$2` |
| Transaction | All-or-nothing group | `pool.Begin` / `Commit` / `Rollback` |
| Tracer | Query timing hooks | New Relic + local logger |

Next: **[Chapter 3: Migrations with Tern](03_migrations_with_tern.md)** — how the schema itself is created and versioned.
