# What the Extra Layers Buy You

Back to the [learning guide](README.md). A small CRUD app is a valid way to learn;
more dependencies do not automatically make an application production-ready.
Tasker adds patterns that are useful to understand along with their costs.

| Choice | Benefit | Cost or current boundary |
| --- | --- | --- |
| Handler/service/repository layers | Separate HTTP, decisions and SQL | More wiring; services still depend on Echo and concrete repositories |
| Clerk | Delegates identity/session machinery | External dependency; ownership checks remain application work |
| pgxpool | Reuses database connections | Capacity and query lifetimes still require tuning; declared limits are not assigned here |
| Tern migrations | Reproducible ordered schema changes | Applied migration history must be preserved; current migrations have no down SQL |
| React Query | Shared server-data cache and mutation lifecycle | Cache keys, invalidation and user-switch handling need deliberate design |
| ts-rest/Zod | Typed frontend calls and reusable schemas | No generated Go backend; runtime and SQL agreement still need tests |
| Redis/Asynq | Work can run separately from user requests | Another service, retries, duplicate effects and operational visibility |
| S3 storage | File bytes separated from relational records | Cross-system consistency, size limits, keys and cleanup |
| Structured logs/APM | Correlate failures and timings | Configuration, telemetry cost and privacy concerns |
| Testcontainers helpers | Isolated realistic PostgreSQL for tests | Docker dependency; no feature coverage until tests actually exist |

## Synchronous and Asynchronous Work

A synchronous function completes before its caller continues. A queued task is
serialized, stored and consumed later. That can avoid making an HTTP request
wait for delivery, but it does not guarantee a specific latency or delivery time.

Here, [cron commands](../../apps/backend/internal/cron/README.md) enqueue
reminders/reports. The welcome-task constructor is not automatically called on
signup. Queue retries can repeat a side effect, so successful enqueue and
exactly-once email delivery are different claims.

## Types and Runtime Behavior

Changing a shared TypeScript shape helps checked client code discover mismatches.
It does not check handwritten SQL or independent Go JSON tags. A schema must be
executed to validate runtime values. Build success is one check, not the whole
verification strategy. Read the [package guide](../../packages/README.md).

## Separation Is Practical, Not Absolute

The code passes a shared server object into many constructors. Services take
Echo contexts, and repositories contain some defaults and status rules. You
cannot swap a database or web framework without considering those dependencies.
That does not make layers useless; it describes the actual degree of coupling.

Before adding another abstraction, name the real duplication or testability
problem it solves. For learning, one well-understood vertical operation is more
valuable than many folders whose responsibilities you cannot explain.
