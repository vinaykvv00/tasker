# Build One Feature From Model to HTTP

Back to [README.md](README.md). First trace the implemented Category creation
path, then use the same sequence to add a small feature yourself.

## Start With One Small Complete Operation

Use **create a category** first. Its repository is smaller than a populated Todo query and teaches the same core pattern. Once the path is tested, repeat the pattern for Todo creation, then learn the extra concepts in listing, hierarchy, comments, and statistics.

Implemented example: `POST /api/v1/categories`, accepting:

```json
{
  "name": "Learning",
  "color": "#22aa88",
  "description": "Backend practice"
}
```

Do not send `user_id`, an ID to assign, or server timestamps. Authentication
provides ownership and PostgreSQL generates the stored identity/timestamps.
Use disposable data when exercising this write endpoint.

## Step 1: Describe the Data and Invariants

Read the category table in [../internal/database/migrations/002_todos.sql](../internal/database/migrations/002_todos.sql), then [../internal/model/category/category.go](../internal/model/category/category.go) and [../internal/model/category/dto.go](../internal/model/category/dto.go).

Ask: which fields are required, which can be NULL, which are generated, and which must be unique? A DTO's limits and a database column's limits need not be identical. In this case, `(user_id, name)` uniqueness belongs to the database, not just a pre-check.

Write input-validation tests first. Verify the actual tags with realistic values, including optional description fields; do not assume every tag inherited from a tutorial is correct.

## Step 2: Verify the Existing Repository

`CreateCategory` already inserts parameterized values and maps the returned row. Use a disposable migrated database to verify the returned ID, owner, name, and timestamps. Then test duplicate names and ownership isolation in detail/update/delete methods.

Use [../internal/repository/README.md](../internal/repository/README.md) for the query mechanics. Do not test repository SQL by adding ad hoc writes to main or by inspecting only startup logs.

## Step 3: Read the Use Case

The [Category service](../internal/service/category.go) delegates creation to
the repository and logs a business event. More involved operations need extra
decisions. Services should not build HTTP headers or duplicate SQL.

For Todo creation, this same step must verify related parent/category ownership. For comments, it must verify access to the referenced Todo. Pass the authenticated identity explicitly. Think about race conditions and database constraints as well as read-before-write checks.

The registry constructs Category, Todo and Comment alongside Auth and Job. A
new feature must be constructed and stored there before handlers can use it.
See [../internal/service/README.md](../internal/service/README.md).

## Step 4: Add the HTTP Boundary

The [Category handler](../internal/handler/category.go) allocates a **fresh DTO
per request**, binds input, validates it, retrieves authenticated identity and
calls the service. It returns 201 on creation and lets the global error handler
format failures. Preserve this pattern in new handlers.

The existing generic helpers centralize much of this, but review their captured-request-pointer behavior and the validation helper's error handling before using them. Avoid sharing a mutable DTO allocated once during route registration.

## Step 5: Connect All Registries

```mermaid
flowchart LR
    Model[Model and DTO] --> Repo[Existing Category repository]
    Repo --> Service[Category service]
    Service --> Services[Services constructor]
    Services --> Handler[Category handler]
    Handler --> Handlers[Handlers constructor]
    Handlers --> Route[Authenticated route]
    Route --> Test[Exercise HTTP behavior]
```

This is implementation order, not the direction an HTTP request travels. At runtime the route calls the handler, then service, then repository.

Add the route under an explicitly authenticated group. Merely having `RequireAuth` available in a registry does not apply it. Do not trust a user ID from request JSON as a substitute.

## Step 6: Verify the Boundary Before Adding More Features

| Case | What to prove |
| --- | --- |
| Missing/invalid token | Authentication rejects the request |
| Bad JSON or invalid fields | Consistent client error, not panic or SQL execution |
| Valid create | Correct success status, JSON shape, and stored owner |
| Duplicate name for same owner | Constraint error translated consistently |
| Same name for different owner | Allowed, according to the schema |
| Detail/update/delete for another owner | No access and no unintended disclosure |
| Concurrent requests | No shared DTO state or data race |

Agree the contract before implementing a new feature, then check both sides as
you build. Category's contract and UI already exist. Check method, URL prefix,
authorization header and response fields in the Network tab. A shared
TypeScript type cannot register an Echo route.

## Then Extend to the More Complex Todo Cases

- A populated Todo needs joins/aggregation and tests for duplicate child/comment rows.
- A list needs validated defaults, an allowlisted sort, properly grouped search filters, stable ordering, and matching counts.
- A parent update needs ownership checks and cycle rules beyond a direct self-parent constraint.
- A partial update needs a deliberate distinction between omitted fields and explicitly clearing nullable fields.
- Multiple atomic database writes need the same transaction; calls on the ordinary pool are outside it.

The [repository guide](../internal/repository/README.md) records remaining
aggregation and update limitations. Do not treat tutorial code as a finished
design. This documentation leaves application fixes to feature work.

## Stop and Explain These Before Moving On

1. Why doesn't creating a model create a table?
2. Why is a repository method not callable directly from the browser?
3. Why are authentication and ownership checks different?
4. What does `%w` preserve for the error handler?
5. Why can Go compile a file whose SQL fails at runtime?
6. What test would catch one user's data appearing in another user's search?

Use the linked folder guide for the concept you cannot yet explain; do not reread every constructor just because its feature name changed.