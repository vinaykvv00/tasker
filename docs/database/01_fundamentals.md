# Chapter 1: Database Fundamentals

Before we touch any Go code, we need a shared vocabulary. This chapter explains what a relational database actually *is*, using plain language and small examples.

---

## 1.1 What is a Database?

A **database** is an organized collection of data. A database management system (DBMS), such as PostgreSQL, is the software that stores it and answers queries.

In this project PostgreSQL runs separately from the Go application. Other systems, such as SQLite, can be embedded in an application. This PostgreSQL server:

- Persists data to disk (so it survives restarts)
- Enforces rules about that data (so it stays correct)
- Answers queries (so your app can read it back)

Think of it as a **filing cabinet that lives in another room**. Your app doesn't hold the files; it asks the cabinet to fetch or store them.

In this project, that cabinet is **PostgreSQL**, running inside a Docker container.

---

## 1.2 Relational Databases and Tables

A **relational database** organizes data into **tables** — like spreadsheets with strict rules.

A **table** has:

- **Columns** — the fields (name + type), fixed in advance
- **Rows** — the individual records

Example — a `todos` table:

| id | title | status | priority | created_at |
| :--- | :--- | :--- | :--- | :--- |
| `a1b2...` | Buy groceries | `draft` | `medium` | 2026-10-03 10:00 |
| `c3d4...` | Write docs | `completed` | `high` | 2026-10-03 11:30 |

The **columns** are `id`, `title`, `status`, `priority`, `created_at`.
The **rows** are the two todos.

---

## 1.3 SQL — The Language of Databases

**SQL** (Structured Query Language) is the main language used here to define tables and query data. PostgreSQL also supports procedural languages such as PL/pgSQL for stored functions. pgx transports commands and values using PostgreSQL's network protocol.

SQL splits into a few families:

| Family | Purpose | Example |
| :--- | :--- | :--- |
| **DDL** (Data Definition) | Define structure | `CREATE TABLE`, `ALTER TABLE`, `DROP TABLE` |
| **DML** (Data Manipulation) | Work with rows | `SELECT`, `INSERT`, `UPDATE`, `DELETE` |
| **DCL** (Data Control) | Permissions | `GRANT`, `REVOKE` |

Your migration files are mostly **DDL**:

```sql
CREATE TABLE todos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL
);
```

Your repositories will be mostly **DML**:

```sql
SELECT id, title FROM todos WHERE user_id = $1;
```

> **Note on `$1`:** That is a **parameter placeholder**. The driver substitutes the real value safely. This is how you avoid SQL injection — never build SQL by string concatenation.

---

## 1.4 Data Types

Every column has a **type** that constrains what it can hold. Common PostgreSQL types used in this project:

| Type | Meaning | Example value |
| :--- | :--- | :--- |
| `UUID` | 128-bit globally unique ID | `a1b2c3d4-...` |
| `TEXT` | Unlimited-length string | `"Buy groceries"` |
| `TIMESTAMPTZ` | Timestamp **with** timezone | `2026-10-03 10:00:00+00` |
| `JSONB` | Binary JSON (queryable) | `{"tags": ["home"]}` |
| `SERIAL` | Auto-incrementing integer | `1`, `2`, `3` |
| `BOOLEAN` | True / false | `true` |

### Why `TIMESTAMPTZ` and not `TIMESTAMP`?

`TIMESTAMP` stores a wall-clock date/time without a timezone. `TIMESTAMPTZ` represents an instant and renders it in the session timezone; it does not retain the original named timezone. Prefer it for these creation/update/deadline instants. Local recurring schedules can additionally need a timezone name and local-time rules.

### Why `UUID` and not `SERIAL` for IDs?

- **Very low collision probability** with random UUIDs; the primary key still enforces uniqueness.
- **Harder to enumerate** than sequential IDs, but never a replacement for ownership checks.
- **Independent generation** without coordinating a shared integer sequence.

---

## 1.5 Keys

A **key** is a column (or set of columns) with a special guarantee.

### Primary Key

Uniquely identifies each row. Cannot be `NULL`. Every table should have one.

```sql
id UUID PRIMARY KEY DEFAULT gen_random_uuid()
```

### Foreign Key

A column that **references** a primary key in another table, enforcing that the referenced row exists.

```sql
category_id UUID REFERENCES todo_categories ON DELETE SET NULL
```

This says: *"`category_id` must match an existing `todo_categories.id` — or be `NULL`."*

The `ON DELETE` clause decides what happens when the parent is deleted:

| Clause | Behavior |
| :--- | :--- |
| `ON DELETE CASCADE` | Delete the child rows too |
| `ON DELETE SET NULL` | Set the child's column to `NULL` |
| `ON DELETE NO ACTION` | Default: reject a remaining foreign-key violation when checked; checks may be deferred if configured |
| `ON DELETE RESTRICT` | Refuse the referenced deletion without deferring this check |

In this project:
- `todo_comments.todo_id` uses **`CASCADE`** — deleting a todo deletes its comments.
- `todos.category_id` uses **`SET NULL`** — deleting a category keeps the todo, just uncategorized.

---

## 1.6 Constraints

Constraints are **rules the database enforces for you**. If a write violates a rule, Postgres rejects it. A `DEFAULT`, listed alongside them below, supplies a value rather than restricting allowed values; explicitly inserting NULL does not invoke a default.

| Constraint | Meaning |
| :--- | :--- |
| `NOT NULL` | Value is required |
| `UNIQUE` | No two rows may share this value |
| `DEFAULT` | Value used when none is supplied |
| `CHECK` | Custom boolean condition |
| `PRIMARY KEY` | `NOT NULL` + `UNIQUE` combined |

Example from this project:

```sql
status TEXT NOT NULL DEFAULT 'draft'
```

- `NOT NULL` → every todo must have a status
- `DEFAULT 'draft'` → if you don't provide one, it becomes `'draft'`

---

## 1.7 Indexes

An **index** is a lookup structure that makes certain queries fast — like the index at the back of a book.

Without a suitable index, PostgreSQL may scan the table to find matches. An index gives its planner another access path; the planner can still prefer a sequential scan for a small table or a query matching many rows.

```sql
CREATE INDEX idx_todos_user_id ON todos(user_id);
```

### Trade-offs

| Pros | Cons |
| :--- | :--- |
| Fast reads | Slower writes (index must be updated) |
| Fast sorting / joins | Extra disk space |

**Rule of thumb:** index columns you frequently filter or join on — not every column.

### Unique Index

A `UNIQUE INDEX` enforces uniqueness *and* speeds up lookups:

```sql
CREATE UNIQUE INDEX todo_categories_unique_name
    ON todo_categories(user_id, name);
```

This is a **composite** unique index: a user cannot have two categories with the same name, but two *different* users can.

---

## 1.8 Triggers and Functions

A **trigger** is code that runs automatically when a table changes. A **function** is reusable logic.

This project defines a function in `001_setup.sql`:

```sql
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
```

And attaches it to tables:

```sql
CREATE TRIGGER set_updated_at_todos
    BEFORE UPDATE ON todos
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();
```

**Result:** every time a `todos` row is updated, `updated_at` is refreshed automatically. Your Go code never has to remember to set it.

---

## 1.9 Transactions (ACID)

A **transaction** groups multiple statements so they succeed or fail **together**.

The classic example: transferring money.

```sql
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;
```

If the second statement fails, the first is rolled back — no money vanishes.

This is summarized by **ACID**:

| Letter | Property | Meaning |
| :--- | :--- | :--- |
| **A** | Atomicity | All or nothing |
| **C** | Consistency | Rules always hold |
| **I** | Isolation | Concurrent transactions don't corrupt each other |
| **D** | Durability | Committed data survives crashes |

Isolation has levels: ordinary transactions do not automatically eliminate every concurrency anomaly. The example above explains atomicity, not a complete payment system. Durability also depends on database/storage configuration and does not replace backups.

`pgx` supports transactions via `pool.Begin(ctx)` — covered in Chapter 2.

---

## 1.10 Schema

The **schema** is the *shape* of your database: all its tables, columns, types, indexes, and constraints.

> **Schema = structure. Data = contents.**

Two databases can have the *same schema* but *different data* (e.g., your local DB vs. production).

The schema is defined by your **migrations** — which is exactly what Chapter 3 covers.

---

## 1.11 Summary

| Concept | One-line meaning |
| :--- | :--- |
| Database | A separate program that stores and serves data |
| Table | A structured collection of rows |
| Column | A typed field in a table |
| Row | A single record |
| SQL | The language used to talk to the database |
| Primary key | Unique row identifier |
| Foreign key | A reference to another table's row |
| Constraint | A rule the database enforces |
| Index | A structure that speeds up lookups |
| Trigger | Code that runs automatically on changes |
| Transaction | An all-or-nothing group of statements |
| Schema | The overall structure of the database |

Next: **[Chapter 2: PostgreSQL and the pgx Driver](02_postgres_and_pgx.md)** — how Go actually talks to this database.
