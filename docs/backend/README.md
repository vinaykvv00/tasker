# Backend Learning Guide

Start at **[../../apps/backend/docs/README.md](../../apps/backend/docs/README.md)** for the ordered beginner guide, diagrams, current feature status, and complete folder/file map.

The goal is to understand the running Go application and trace a feature through
its layers. Detailed docs live beside the code so you can read a folder's
explanation while working in it.

## Read in Stages

| Stage | Guide |
| --- | --- |
| Learn the Go syntax used here | [../../apps/backend/docs/go-reading-basics.md](../../apps/backend/docs/go-reading-basics.md) |
| Follow startup and shared resources | [../../apps/backend/internal/server/README.md](../../apps/backend/internal/server/README.md) |
| Understand models, DTOs, and field tags | [../../apps/backend/internal/model/README.md](../../apps/backend/internal/model/README.md) |
| Read one CRUD pattern, then joins and lists | [../../apps/backend/internal/repository/README.md](../../apps/backend/internal/repository/README.md) |
| Learn what the service layer should decide | [../../apps/backend/internal/service/README.md](../../apps/backend/internal/service/README.md) |
| Connect a feature end to end | [../../apps/backend/docs/building-a-feature.md](../../apps/backend/docs/building-a-feature.md) |
| Run and inspect the local system | [../../apps/backend/docs/local-development.md](../../apps/backend/docs/local-development.md) |
| Verify it without touching valuable data | [../../apps/backend/internal/testing/README.md](../../apps/backend/internal/testing/README.md) |

For SQL, PostgreSQL, pgx, and Tern from first principles, use [../database/README.md](../database/README.md).

## Current Implementation, Not a Finished Architecture

The backend starts from [../../apps/backend/cmd/tasker/main.go](../../apps/backend/cmd/tasker/main.go). It constructs shared resources and then repository, service, handler, and router registries.

Todo, Category and Comment models, repositories, services, handlers and routes
are wired. Attachments, statistics, background email jobs and a separate cron
entry point are present. React now has authentication, dashboard, todo/category
screens and settings.

The folder guides label planned flows and known code limitations. A startup log, a model definition, or a helper file existing is not proof of a working, authorized feature. Application code is intentionally unchanged by this documentation work.

## Important Boundaries

Local startup does not auto-run migrations. Configuration uses case-sensitive,
dotted names such as `TASKER_DATABASE.HOST`. Features live under `/api/v1`;
`/status` is separate. Test helpers are infrastructure, not proof of feature
coverage, and deployment remains a later learning stage.

Use the [end-to-end guide](../overall_end_to_end/README.md) to connect backend
code to browser behavior and TypeScript contracts.
