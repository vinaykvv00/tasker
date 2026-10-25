# Routing: Which Code Receives a Request?

Back to [../../docs/README.md](../../docs/README.md).

| File | Responsibility |
| --- | --- |
| [router.go](router.go) | Creates Echo, assigns the global error handler, installs global middleware, and groups routes |
| [system.go](system.go) | Registers currently available infrastructure endpoints |

## Current Routes

| Method/path | Handler or behavior |
| --- | --- |
| `GET /status` | `Health.CheckHealth` |
| `GET /docs` | `OpenAPI.ServeOpenAPIUI` |
| `/static/*` | Files from the backend's static directory, including the OpenAPI JSON |

There is no root `/` page, so browsing port 8080 at `/` can correctly return 404. The frontend is a separate application, not HTML automatically served by the Go backend.

`router.Group("/api/v1")` currently creates an empty prefix group. It does not generate CRUD endpoints, register all handlers, or attach authentication. The `services` parameter is also unused in the current router constructor.

## Read a Registration

```go
r.GET("/status", h.Health.CheckHealth)
```

This tells Echo to invoke the method value `h.Health.CheckHealth` for that method/path combination. It does not invoke the method immediately. Startup stores the handler; a later matching HTTP request calls it.

A path such as `/api/v1/todos/:id` would capture an ID from the URL. This is a hypothetical route, not one currently registered. The method also matters: registering GET does not register POST, PATCH, or DELETE.

## Middleware and Authentication

The router installs rate limiting, CORS, security headers, request IDs, tracing, context enrichment, logging, and panic recovery. Global middleware can affect requests even when no feature route matches.

Although an Auth middleware object is constructed, `RequireAuth` is not attached here. When implementing feature routes, explicitly apply it to the relevant group/routes before relying on a user ID. Initializing Clerk's key alone does not protect an endpoint.

## Frontend Connection, Briefly

The browser runs the frontend and makes HTTP requests to the backend origin. Origin includes scheme, hostname, and port, so ports 3000 and 8080 are different origins. CORS controls browser access to cross-origin responses; it is not authentication and does not prevent direct HTTP clients from sending requests.

Shared TypeScript contracts describe expected API shapes, but they do not generate this Go route table or prove the Go implementation follows them. Add the backend handler, route, and tests as well as the client contract.

Follow the request into [../handler/README.md](../handler/README.md).