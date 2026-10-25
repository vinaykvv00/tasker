# Handlers: The HTTP Boundary

Back to [../../docs/README.md](../../docs/README.md).

A handler understands HTTP: route parameters, JSON bodies, query strings, response status, and headers. For a feature, it should bind/validate input, obtain trusted identity, call the service, and write the result. It should not duplicate repository SQL.

## What Exists

| File | Purpose |
| --- | --- |
| [handlers.go](handlers.go) | Registry containing only Health and OpenAPI handlers |
| [health.go](health.go) | Directly checks the DB and Redis and returns JSON |
| [openapi.go](openapi.go) | Reads the local OpenAPI HTML file and returns HTML |
| [base.go](base.go) | Shared handler state and generic wrappers for JSON, no-content, and file responses |

There are no Tasker feature handlers yet. `NewHandlers` receives `services` but does not currently use it. A feature needs both a handler and explicit route registration.

## First Example: CheckHealth

`HealthHandler` embeds `Handler`, which holds the shared server pointer. `CheckHealth(c echo.Context) error` matches Echo's handler signature.

It builds a map containing `status`, `timestamp`, `environment`, and `checks`. It pings the database and, when a client exists, Redis. `c.JSON(status, response)` serializes the Go value and writes the HTTP response.

Actual top-level status strings are `healthy` and `unhealthy`, not `UP`. Check results are nested under `checks`, not `details`.

Current limitation: database failure sets the overall `isHealthy` flag to false, but Redis failure only marks its individual check unhealthy. Therefore HTTP 200/top-level healthy does not guarantee Redis is healthy. The check uses fresh background contexts with five-second timeouts, not the request's cancellation context.

This infrastructure endpoint legitimately checks server dependencies directly; it is not the template for putting Todo business logic in a handler.

## Generic Helpers Without the Repetition

Read `Handle`, `HandleNoContent`, and `HandleFile` as variations of one process:

```mermaid
flowchart LR
    Request[Echo request] --> Bind[BindAndValidate]
    Bind --> Callback[Typed handler callback]
    Callback --> Response[Selected response writer]
    Bind -. error .-> Errors[Global HTTP error handler]
    Callback -. error .-> Errors
```

`HandlerFunc[Req, Res]` is a function type: it receives an Echo context and a validated request value, and returns a typed result plus an error. `Req` must satisfy `Validatable`; `Res any` allows different response types.

`handleRequest` contains the common work: timing, request-scoped logging, optional tracing, validation, callback execution, and response writing. The `ResponseHandler` interface lets JSON/no-content/file implementations supply the final response behavior without repeating that flow.

| Helper | Expected response |
| --- | --- |
| `Handle` | JSON model or response object |
| `HandleNoContent` | HTTP status with no response body |
| `HandleFile` | Byte content with filename and content-type headers |

Returning an error lets Echo invoke the configured global error handler. Avoid writing a success body and then returning an unrelated error; once the response is committed, its status/body cannot simply be replaced.

## Important Detail Before Using These Helpers

The helpers currently close over the supplied `req` value. If a pointer is allocated once at route registration, requests reuse that same object: fields can leak between requests and concurrent binding can race. Each request needs a fresh mutable DTO. Treat the existing wrappers as code to review before wiring new feature routes, not proof that request allocation is already safe.

For a future Todo handler, the checklist is: fresh DTO, successful bind/validation/defaults, authenticated user ID, request-scoped Go context, service call, correct success status, returned errors, and tests for invalid input and another user's resource.

For the binding implementation, see [../validation/README.md](../validation/README.md). For the missing business layer, see [../service/README.md](../service/README.md).