# Middleware: Work Around a Handler

Back to [../../docs/README.md](../../docs/README.md).

Middleware wraps an HTTP handler. It may inspect a request, reject it, attach context, call the next handler, and observe the response/error. A middleware declaration is not active until the router attaches it.

## File Map

| File | Responsibility |
| --- | --- |
| [middlewares.go](middlewares.go) | Constructs the middleware registry from shared server resources |
| [global.go](global.go) | CORS, security headers, recovery, request logging, global error translation |
| [auth.go](auth.go) | Clerk authorization wrapper and extraction of identity/role/permissions |
| [request_id.go](request_id.go) | Request-ID middleware and lookup helper |
| [context.go](context.go) | Request-scoped logger, identity lookup, and trace enrichment |
| [tracing.go](tracing.go) | Optional New Relic middleware and extra trace attributes |
| [rate_limit.go](rate_limit.go) | Records rate-limit events; the actual limiter is installed in the router |

## The Wrapper Shape

The conceptual shape is `func(next echo.HandlerFunc) echo.HandlerFunc`: accept a handler and return a wrapped one. Returning early rejects the request; calling `next(c)` continues processing. Code after `next(c)` can measure duration or inspect an error.

Middleware order matters. A logger created before authentication cannot initially contain identity that authentication has not yet discovered. Recovery also only covers code inside its wrapper; its presence is not a guarantee that every earlier middleware is protected.

## Authentication vs. Authorization

`RequireAuth` wraps Clerk's header authorization, obtains session claims from the request's standard context, and writes `user_id`, `user_role`, and `permissions` into Echo's context. The subject is the authenticated identity.

This authenticates the caller. It does not by itself prove that a particular Todo, category, or comment belongs to them, and storing roles/permissions does not enforce a role policy. Services and user-scoped SQL still need authorization checks.

At this snapshot [../router/router.go](../router/router.go) constructs Auth but does not attach `RequireAuth` to any feature route. The `/api/v1` group is empty.

## Two Contexts, Not One

| Context | Typical content and use |
| --- | --- |
| `echo.Context` | Request/response helpers, `c.Set` / `c.Get`, bound route parameters |
| `context.Context` from `c.Request().Context()` | Cancellation/deadlines and values passed to pgx, tracing, or other Go APIs |

`ContextEnhancer` creates a logger containing request fields and stores it in both places. Identity set with `c.Set` is not automatically a value in the standard context. `GetUserID(c)` reads the Echo value and returns an empty string if absent.

`GetLogger(c)` falls back to a no-op logger when context enrichment has not installed one. No output from that logger does not necessarily mean no code executed.

## Logs, Errors, and Security Headers

`RequestLogger` produces an API summary including URI, method, latency, status, and request context. `GlobalErrorHandler` separately logs the original failure and formats a response. Two log entries can therefore describe one failed request.

CORS is a browser-origin policy, not login. Security headers are useful defaults, not a complete security policy. The in-memory rate limiter operates per application process; it is not a Redis-backed global limit shared across replicas.

For the error journey and response fields, see [../errs/README.md](../errs/README.md).