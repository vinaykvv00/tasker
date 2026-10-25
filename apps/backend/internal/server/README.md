# Server Resources and Application Lifetime

Back to [../../docs/README.md](../../docs/README.md). Read [server.go](server.go) beside [../../cmd/tasker/main.go](../../cmd/tasker/main.go).

## Server Does Not Mean Only HTTP

`Server` groups shared resources: configuration, logger, optional New Relic application, database pool, Redis client, background-job service, and a private `http.Server`.

Most of these are constructed once and reused by many requests. Passing `*server.Server` to a repository is manual dependency injection: the constructor receives an existing dependency instead of reading global settings and opening another connection.

## Startup, in Actual Order

```mermaid
sequenceDiagram
    participant Main as main
    participant Config as config
    participant Migration as database.Migrate
    participant Server as server.New
    participant Resources as PostgreSQL / Redis / Asynq
    participant Wiring as registries and router
    Main->>Config: LoadConfig
    Main->>Main: Initialize logger service and logger
    opt environment is not local
        Main->>Migration: Apply embedded migrations
    end
    Main->>Server: New(cfg, logger, loggerService)
    Server->>Resources: Connect DB; ping Redis; start job worker
    Server-->>Main: Shared Server
    Main->>Wiring: Repositories -> Services -> Handlers -> Router
    Main->>Server: SetupHTTPServer(router)
    Main->>Server: Start in a goroutine
    Main->>Main: Wait for interrupt
```

`database.Migrate` and `database.New` are different operations. The former changes schema; the latter prepares a pool for normal queries. Local startup skips the first, not the second.

## Read the Server Methods

| Method | What it does |
| --- | --- |
| `New` | Creates DB, constructs/pings Redis, initializes email-backed job handlers, and starts Asynq |
| `SetupHTTPServer` | Attaches the Echo router as an `http.Handler` and sets address/timeouts |
| `Start` | Checks setup happened, logs the intended port, then blocks in `ListenAndServe` |
| `Shutdown` | Stops HTTP, closes the DB pool, then stops the job service |

Echo satisfies the standard library's `http.Handler` interface, so `http.Server` can invoke the router for each request. This is an example of Go's implicit interfaces.

`Addr: ":" + port` listens on the available interfaces, not exclusively localhost. Network exposure also depends on the host/firewall setup.

## What Startup Logs Prove

- `connected to the database` means the pool's startup ping succeeded. It does not mean all migrations or feature queries have been tested.
- Redis ping failure is logged and startup continues. Asynq is still started and may repeatedly fail to access Redis.
- `starting server` is logged before `ListenAndServe` attempts to bind. A later port-in-use error is still possible.
- No compile error does not mean SQL strings, DTO tags, or external credentials work.

## Interrupt and Shutdown

`signal.NotifyContext` listens for `os.Interrupt` in the current main function. Main waits on `ctx.Done()`, then creates a fresh background context with a 30-second timeout for shutdown. It does not reuse the already-canceled signal context.

`http.ErrServerClosed` is expected after normal HTTP shutdown, so main excludes it from fatal startup errors.

Current limitations: main explicitly registers interrupt, not SIGTERM; the standalone Redis client is not explicitly closed; shutdown closes the DB before stopping jobs; and an early shutdown error skips later cleanup. Constructor failures also deserve cleanup tests. These are lifecycle follow-ups, not guarantees supplied automatically by Echo or Go.

## Resource Lifetime vs. Request Lifetime

The shared server survives across requests. Each handler receives a separate request context and input. pgxpool manages concurrent connections; a borrowed connection is held only while the query/rows/transaction uses it. Do not store per-user state in shared server fields.