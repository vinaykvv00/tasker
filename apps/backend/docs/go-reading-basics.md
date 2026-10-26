# Reading the Go in This Backend

[README.md](README.md) is the starting point. This is a vocabulary guide, not a second implementation of the application.

## Packages, Files, and Imports

A directory normally contains one Go package. All its ordinary Go files compile together; the filename does not create a separate class. For example, the repository registry and the feature repositories all declare `package repository`.

- [../go.mod](../go.mod) declares the module's import path and dependency versions.
- [../go.sum](../go.sum) contains dependency checksums, not installed executables.
- `github.com/vinaykvv00/go-tasker/internal/model/todo` refers to a package in this module. The module name can differ from the local folder name `tasker`.
- `internal` restricts which other Go packages may import this code. It is not an HTTP access-control mechanism.
- `package main` plus `func main()` makes the executable entry point.
- Capitalized identifiers such as `TodoRepository` are exported; lowercase fields such as `server` are package-private.

An import alias, such as `loggerPkg`, changes the name used in that file, not the underlying package.

## Read a Method Signature

From [../internal/repository/category.go](../internal/repository/category.go):

```go
func (r *CategoryRepository) GetCategoryByID(ctx context.Context, userID string, categoryID uuid.UUID) (*category.Category, error)
```

This is a signature excerpt, not a complete function to paste into a Go file.

| Part | Meaning |
| --- | --- |
| `func` | Declares a function or method |
| `(r *CategoryRepository)` | Receiver: this method belongs to a repository value; `r` refers to it |
| `ctx context.Context` | Carries cancellation, deadlines, and request-scoped values |
| `userID string` | Explicit identity input; it must come from trusted authentication |
| `categoryID uuid.UUID` | Typed identifier of the requested resource |
| `(*category.Category, error)` | Two results: a pointer to data, and an error |

`context.Context` is from Go's standard library, not Echo. An HTTP handler can obtain one from `c.Request().Context()`. Echo's `echo.Context` is a different object that also provides binding and response-writing methods.

## Structs, Pointers, and Constructors

```go
type CategoryRepository struct {
    server *server.Server
}
```

A struct groups fields. `*server.Server` is a pointer to a shared server value. It does not create a new database connection each time the repository is used.

```go
return &CategoryRepository{server: server}
```

The braces construct the struct; `&` returns its address. `NewCategoryRepository` is an ordinary function following a naming convention, not a language-level constructor or framework hook.

Three related symbols have different roles:

| Expression | Meaning |
| --- | --- |
| `*string` | A pointer type |
| `&name` | Address of an existing value |
| `*payload.Name` | Read the pointed-to value; requires a non-nil pointer |

`nil` means no value for pointers, slices, maps, interfaces, and some other types. A pointer field is useful for optional input, but dereferencing nil panics. A plain `bool` cannot distinguish an omitted input from `false`; `*bool` can, with the JSON-null caveat in the model guide.

## Assignment, Maps, and Slices

- `value := expression` declares a local variable with an inferred type.
- `value = expression` assigns an existing variable.
- `var total int` starts at Go's zero value, which is `0` for integers.
- `pgx.NamedArgs{"user_id": userID}` is a map of parameter names to values.
- `[]string{}` is a slice of strings, used for variable-length SQL conditions.
- `append(conditions, condition)` returns a slice with an additional element.
- `strings.Join(conditions, " AND ")` combines conditions; it does not validate SQL.
- Backticks delimit raw strings, useful for multiline SQL. Struct tags also use backtick strings, interpreted by libraries through reflection.

## Errors Are Return Values

```go
if err != nil {
    return nil, fmt.Errorf("failed to query category: %w", err)
}
```

This illustrative fragment returns early on failure. `%w` preserves the original error so another layer can use `errors.Is` or `errors.As`. Do not decide what an error means by searching its message text.

`return &categoryItem, nil` means success. `return nil, err` means failure. `panic` is not the normal way this project reports invalid client input.

## Embedding, Generics, and Interfaces

- `model.Base` embedded in `Todo` promotes common ID/timestamp fields. This is composition, not class inheritance, and it does not create SQL columns.
- `model.PaginatedResponse[category.Category]` instantiates a generic response with category rows as its data type.
- `pgx.RowToStructByName[category.Category]` selects the Go type the row mapper should construct.
- An interface describes behavior. A type satisfies it implicitly by implementing the required methods. The binding helper uses this for payloads with `Validate() error`.

## Lifetimes and Concurrency

`defer cleanup()` schedules cleanup when the surrounding function returns. It is commonly used for a transaction rollback, rows close, or context cancel. Fatal process exits do not run deferred functions.

`go func() { ... }()` starts a goroutine. In [../cmd/tasker/main.go](../cmd/tasker/main.go), it runs the HTTP listener while the main goroutine waits for an interrupt. `<-ctx.Done()` blocks until that context is canceled.

Shared dependencies live for the application's lifetime. DTOs and request contexts belong to one request. Do not store a request context or one user's mutable payload on a shared repository.

Next: [../internal/model/README.md](../internal/model/README.md).