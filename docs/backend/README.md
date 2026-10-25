# Backend Learning Guide

Start at **[../../apps/backend/docs/README.md](../../apps/backend/docs/README.md)** for the ordered beginner guide, diagrams, current feature status, and complete folder/file map.

The goal is to understand the cloned Go boilerplate first, then connect the Tasker models and repositories to services and HTTP handlers. The detailed docs live beside the backend code so you can read a folder's explanation while working in it.

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

Todo, Category, and Comment models/repositories exist. Feature services and HTTP handlers/routes are not yet wired; Auth and Job are the current service entries, and Health/OpenAPI are the current handler entries. The frontend remains a minimal Hello World screen.

The folder guides label planned flows and known code limitations. A startup log, a model definition, or a helper file existing is not proof of a working, authorized feature. Application code is intentionally unchanged by this documentation work.

## What Changed From the Earlier Overview

Local startup does not auto-run migrations; configuration uses dotted keys such as `tasker_DATABASE.HOST`; test helpers provision PostgreSQL, not Redis; and `/api/v1` is currently an empty group. The new guides use actual code behavior rather than the generic capabilities a boilerplate might eventually provide.
