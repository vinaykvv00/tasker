# Add One Feature Across the Necessary Layers

Back to the [learning index](README.md). This is a workflow for future changes,
not another implemented resource. Start with the existing
[Category walkthrough](../../apps/backend/docs/building-a-feature.md) before
inventing a new architecture or copying an entire feature.

## 1. Define Behavior Before Files

Choose one operation and write examples of success, invalid input, missing
resource and another user's resource. State the HTTP method/path, editable
fields, generated fields, response status and data shape. Decide whether a
change affects existing records or requires backfill.

For a small learning exercise, add or fix one filter/form behavior first. For a
new stored field, trace the full contract below. Not every feature needs a new
table, repository, service or page.

```mermaid
flowchart LR
    Behavior[Acceptance examples] --> Contract[Agree request and response]
    Contract --> Data[Schema and Go data shapes if needed]
    Data --> Backend[Repository and service behavior]
    Backend --> HTTP[Handler and authenticated route]
    HTTP --> UI[Query hook and usable UI]
    UI --> Verify[Focused checks and artifact review]
```

## 2. Use the Closest Existing Pattern

| Concern | Existing reference | What to check |
| --- | --- | --- |
| Schema | [Todo migration](../../apps/backend/internal/database/migrations/002_todos.sql) | Nullability, defaults, constraints, indexes, delete behavior |
| Go types | [Category DTO](../../apps/backend/internal/model/category/dto.go) | JSON/path/query tags, validation, defaults |
| SQL | [Category repository](../../apps/backend/internal/repository/category.go) | Parameters, owner filter, row mapping, wrapped errors |
| Business decisions | [Todo service](../../apps/backend/internal/service/todo.go) | Related ownership, sequencing, partial failures |
| HTTP | [Category handler](../../apps/backend/internal/handler/category.go) | Fresh DTO, trusted identity, status, returned errors |
| Routes | [Category routes](../../apps/backend/internal/router/v1/category.go) | Prefix and `RequireAuth` |
| Contract | [Category contract](../../packages/openapi/src/contracts/category.ts) | Method, path, body, statuses, version prefix |
| Hook | [Category query hooks](../../apps/frontend/src/api/hooks/use-category-query.ts) | Query keys, status checks, mutation invalidation |
| Form | [Category create form](../../apps/frontend/src/components/categories/category-create-form.tsx) | Validation, pending state and payload mapping |

Reuse patterns, not their known defects. Read each layer's caveats before copying
a complex populated query or nullable update builder.

## 3. Change the Database Deliberately

If storage must change, add a **new** migration; do not edit already-applied SQL.
From the backend directory:

```bash
task migrations:new name=describe_the_schema_change
```

Review SQL, target database and data-loss implications before applying:

```bash
task migrations:up TASKER_DB_DSN="postgres://postgres@localhost:5432/tasker?sslmode=disable"
```

That example is only for the current passwordless local database. Never paste a
remote password into history or docs. Use the Tern separator produced by its
template, not an invented separator. Current migrations are forward-only;
rollback is not available merely because a command supports moving backward.

For a new column, update Go row mapping too. `RETURNING *` and strict
`RowToStructByName` can require a matching model field. A migration succeeding
does not prove repository mapping still works.

## 4. Implement Go Behavior

Keep SQL values in parameters. Concatenated sort identifiers must come from an
allowlist. Pass request context and trusted identity to data access; include
ownership predicates in mutations, not only an earlier lookup.

Choose omission/null semantics deliberately. A plain pointer cannot distinguish
all absent-versus-null cases. If clearing a nullable field is the feature,
define presence tracking or a clear operation across every boundary.

For a new resource, connect these registries explicitly:

1. [Repositories](../../apps/backend/internal/repository/repositories.go)
2. [Services](../../apps/backend/internal/service/services.go)
3. [Handlers](../../apps/backend/internal/handler/handlers.go)
4. [Versioned routes](../../apps/backend/internal/router/v1/v1.go)

Existing resources usually need only a new method/route within their current
registrations. Use the shared error pipeline; do not send raw `err.Error()` as a
public 500 response or assert `c.Get("user_id").(string)` without the established
identity boundary.

SQL writes are atomic only when all use the same transaction. Pool queries do
not join a transaction just because a service started one. Redis/S3 side effects
need their own retry/consistency strategy.

## 5. Keep TypeScript Contracts Aligned

Update the relevant [Zod schemas](../../packages/zod/README.md),
[contract](../../packages/openapi/README.md) and exports. Use camelCase JSON
names matching Go JSON tags, including nullability and dates. Versioned feature
contracts use `/v1`; the client adds `/api`.

The namespace matters: calls look like `api.Category.createCategory(...)`, not
`api.createCategory(...)`. Security metadata documents auth; attach Go middleware
separately. TypeScript contracts do not generate or typecheck Go implementation.

From the root, build dependencies and deliberately regenerate API docs:

```bash
bun run build --filter=@tasker/openapi
bun run --cwd packages/openapi gen
```

These are distinct commands. Root build does not run the generator. Review both
JSON artifacts after `gen`; the contract alone does not update the Go static file.

## 6. Build Browser Behavior

Use the existing query hook conventions. Include result-affecting filters in
query keys, check expected statuses, and invalidate affected cached representations
on mutation success. Consider user identity in cache lifetimes.

Cover loading, empty, error, pending and success states. A button needs a working
callback. Normalize select sentinel values before sending them as IDs, serialize
dates deliberately and keep file requests separate from JSON payloads. A
client-side rule is not backend enforcement.

For partial success, show that truth instead of inviting a repeated successful
write. Todo creation followed by uploads is the existing example to study.

## 7. Verify the Smallest Relevant Slice

| Change | Useful check |
| --- | --- |
| DTO/defaults | Valid, invalid, omitted and null input cases |
| SQL/schema | Disposable migrated PostgreSQL; ownership, constraints and multiple related rows |
| Handler/route | HTTP method, auth, status and response shape |
| Contract | Built exports, generated spec and actual Go response agreement |
| Frontend | `tsc -b`, lint and a browser action with Network/Query Devtools |
| External effects | Controlled fake/test provider and partial-failure cases |

There are currently no backend test files; write focused tests as you reach that
learning stage. A command reporting no tests is not coverage. Use
[test helpers](../../apps/backend/internal/testing/README.md) only on disposable data.

From backend: `go build ./...` and `go vet ./...` check compilation/static issues.
From frontend: `bunx tsc -b --pretty false` checks referenced projects. Plain
`tsc` and Vite bundling alone do not establish full type correctness here.

## 8. Review and Commit by Behavior

Review `git diff` and ensure no credentials, generated dependency folders or
reference checkout were added. Keep source and required artifacts together.
Split independent work into meaningful commits, but avoid importing files only
introduced by a later commit.

Use short lowercase messages such as `fix(frontend): reset page after filtering`
or `feat(todos): support clearing due dates`. Describe behavior, not merely
"sync files". Commit and push only when you deliberately choose to.

Final exercise: explain what changed, why each layer needed it, what proves it
works, and what remains unverified.