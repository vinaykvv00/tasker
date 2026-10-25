# Translating PostgreSQL Errors

Read [../errs/README.md](../errs/README.md) for the overall error flow.

| File | What to read for |
| --- | --- |
| [error.go](error.go) | Internal error/severity types, SQLSTATE mapping, and wrapped driver errors |
| [handler.go](handler.go) | Extracting pgx errors and choosing an application error/code/message |

PostgreSQL errors include a SQLSTATE code and details such as table, column, and constraint names. Classifying these fields is more dependable than matching an English error message.

## Current Translation Rules

| Database condition | Current HTTP result |
| --- | --- |
| Foreign-key violation (`23503`) | Bad request, 400 |
| Unique violation (`23505`) | Bad request, 400, not 409 in this implementation |
| Not-null violation | Bad request with a field error |
| Check-constraint violation | Bad request |
| Wrapped `pgx.ErrNoRows` or `sql.ErrNoRows` | Not found, 404 |
| Other/unrecognized failures | Internal server error, 500 |

`errors.As` finds a `*pgconn.PgError` inside wrapped errors. `ConvertPgError` preserves its fields in a local type. Helpers generate a code and human-readable message before constructing `errs.HTTPError`.

## One Example

Two requests try to create the same category name for the same user. The unique index rejects one write. PostgreSQL reports a unique violation; the repository wraps it with `%w`; this translator identifies it; the global handler returns a bad request. A service pre-check alone would not protect against the race, but the database constraint does.

## Limitations

Generated entity names and unique-column messages are heuristic: they depend on table/constraint naming conventions. The not-found path also tries to extract a `table:` marker from the wrapped message. Treat these messages as implementation details to test, not a guaranteed domain vocabulary.

Importing `database/sql` here is for recognizing its no-row sentinel. Normal application queries still use native pgx; this file does not introduce a second database connection layer.