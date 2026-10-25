# How to Build and Add New Features: End-To-End Developer Playbook

This playbook provides a detailed, step-by-step walkthrough for building a new feature—from declaring validation rules all the way to writing database migrations, routing logic on the backend, and API integrations with React UI on the frontend.

We will illustrate this using a hypothetical feature: **A "User Projects" resource** (allowing fetching and creating Projects for an authenticated user).

---

## The Feature Lifecycle Checklist

An end-to-end feature in this monorepo typically flows through five sequential stages:

```
┌─────────────────────────────────────────────────────────┐
│ 1. Schema & Contracts (packages/zod, packages/openapi)  │
└────────────────────────────┬────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────┐
│ 2. Backend Migration (apps/backend Command/Tern script) │
└────────────────────────────┬────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────┐
│ 3. Backend Implementation (Model → Repo → Service → Hnd)│
└────────────────────────────┬────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────┐
│ 4. Backend Router Wiring (apps/backend router config)   │
└────────────────────────────┬────────────────────────────┘
                             ▼
┌─────────────────────────────────────────────────────────┐
│ 5. Frontend Client Hook & View (apps/frontend React JS) │
└─────────────────────────────────────────────────────────┘
```

---

## Step 1: Define Schemas and API Contracts

This project leverages TS-Rest for contract-first design. Writing the API contract first guarantees that our frontend and backend speak the exact same language.

### A. Declare Validation Rules in packages/zod
Create a new schemas file or append validations inside [packages/zod/src/health.ts](packages/zod/src/health.ts) (or create a new file like `packages/zod/src/project.ts` if adding multiple rules).

For our example, register a scheme in [packages/zod/src/health.ts](packages/zod/src/health.ts) or similar:

```typescript
import { z } from "zod";

export const ZProject = z.object({
  id: z.string().uuid(),
  name: z.string().min(3).max(100),
  description: z.string().optional(),
  userId: z.string(),
  createdAt: z.string().datetime(),
});

export const ZCreateProjectSchema = z.object({
  name: z.string().min(3).max(100),
  description: z.string().optional(),
});
```

Export these schemas in the entry-point file [packages/zod/src/index.ts](packages/zod/src/index.ts):
```typescript
export * from "./project.js"; // Ensure the newly created schema file is exported
```

### B. Map Endpoints in API Contracts
Create or update contract declarations inside [packages/openapi/src/contracts](packages/openapi/src/contracts). For example, create `packages/openapi/src/contracts/project.ts`:

```typescript
import { initContract } from "@ts-rest/core";
import { z } from "zod";
import { ZProject, ZCreateProjectSchema } from "@tasker/zod";

const c = initContract();

export const projectContract = c.router({
  getProjects: {
    summary: "Get user projects",
    path: "/projects",
    method: "GET",
    responses: {
      200: z.array(ZProject),
    },
  },
  createProject: {
    summary: "Create a new project",
    path: "/projects",
    method: "POST",
    body: ZCreateProjectSchema,
    responses: {
      201: ZProject,
    },
  },
});
```

Mount `projectContract` inside the master contract list in [packages/openapi/src/contracts/index.ts](packages/openapi/src/contracts/index.ts):

```typescript
import { projectContract } from "./project.js";

export const apiContract = c.router({
  Health: healthContract,
  Project: projectContract, // Wired perfectly!
});
```

### C. Run Codegen to build packages
From the root of the project, run:
```bash
bun build
```
This script validates your types, compiles contracts, and outputs a static OpenApi documentation file to [apps/backend/static/openapi.json](apps/backend/static/openapi.json).

---

## Step 2: Write Database Migration

We store database structures in standard tables. Follow this process to update PostgreSQL safely.

### A. Generate SQL migration script
Change directories into the backend of your project and run Tern:
```bash
cd apps/backend
task migrations:new name=create_projects_table
```
This generates a up/down dual migration template in [apps/backend/internal/database/migrations](apps/backend/internal/database/migrations).

### B. Append SQL Rules
Open your new `.sql` file and write your transaction queries:

```sql
-- Write your Migrate Up queries here
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_projects_user_id ON projects(user_id);

---- create above || put select migration split below ----

-- Write your Migrate Down queries here
DROP TABLE IF EXISTS projects;
```

### C. Run Migration locally
```bash
task migrations:up
```

---

## Step 3: Implement Backend Logic Layers

Now, implement clean architecture components inside [apps/backend/internal/](apps/backend/internal/).

### A. Models Layer
Add structural representation inside [apps/backend/internal/model/](apps/backend/internal/model/) under `project.go`:

```go
package model

import "time"

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UserID      string    `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}
```

### B. Repository Layer (Data Interceptor)
Create [apps/backend/internal/repository/project.go](apps/backend/internal/repository/project.go) to perform Postgres queries:

```go
package repository

import (
	"context"
	"github.com/sriniously/go-tasker/internal/model"
	"github.com/sriniously/go-tasker/internal/server"
)

type ProjectRepository interface {
	Create(ctx context.Context, name, description, userID string) (*model.Project, error)
	GetByUserID(ctx context.Context, userID string) ([]model.Project, error)
}

type projectRepository struct {
	s *server.Server
}

func NewProjectRepository(s *server.Server) ProjectRepository {
	return &projectRepository{s: s}
}

func (r *projectRepository) Create(ctx context.Context, name, description, userID string) (*model.Project, error) {
	query := `INSERT INTO projects (name, description, user_id) VALUES ($1, $2, $3) RETURNING id, name, description, user_id, created_at`
	var p model.Project
	err := r.s.DB.Pool.QueryRow(ctx, query, name, description, userID).Scan(&p.ID, &p.Name, &p.Description, &p.UserID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *projectRepository) GetByUserID(ctx context.Context, userID string) ([]model.Project, error) {
	query := `SELECT id, name, description, user_id, created_at FROM projects WHERE user_id = $1`
	rows, err := r.s.DB.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.UserID, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}
```

Register your repository implementation within [apps/backend/internal/repository/repositories.go](apps/backend/internal/repository/repositories.go).

### C. Services Layer (Business Rules)
Create [apps/backend/internal/service/project.go](apps/backend/internal/service/project.go) to hold processes (like filtering data or firing notification emails via queue workers):

```go
package service

import (
	"context"
	"github.com/sriniously/go-tasker/internal/model"
	"github.com/sriniously/go-tasker/internal/repository"
	"github.com/sriniously/go-tasker/internal/server"
)

type ProjectService interface {
	CreateProject(ctx context.Context, name, description, userID string) (*model.Project, error)
	GetUserProjects(ctx context.Context, userID string) ([]model.Project, error)
}

type projectService struct {
	s    *server.Server
	repo repository.ProjectRepository
}

func NewProjectService(s *server.Server, repo repository.ProjectRepository) ProjectService {
	return &projectService{s: s, repo: repo}
}

func (svc *projectService) CreateProject(ctx context.Context, name, description, userID string) (*model.Project, error) {
	// Add business rule checks: e.g., restrict max projects per user
	return svc.repo.Create(ctx, name, description, userID)
}

func (svc *projectService) GetUserProjects(ctx context.Context, userID string) ([]model.Project, error) {
	return svc.repo.GetByUserID(ctx, userID)
}
```

Register your service implementation within [apps/backend/internal/service/services.go](apps/backend/internal/service/services.go).

### D. Handlers Layer (HTTP Adapter)
Create [apps/backend/internal/handler/project.go](apps/backend/internal/handler/project.go):

```go
package handler

import (
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/sriniously/go-tasker/internal/server"
	"github.com/sriniously/go-tasker/internal/service"
)

type ProjectHandler struct {
	Handler
	service service.Services
}

type CreateProjectInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewProjectHandler(s *server.Server, svc *service.Services) *ProjectHandler {
	return &ProjectHandler{
		Handler: NewHandler(s),
		service: *svc,
	}
}

func (h *ProjectHandler) GetProjects(c echo.Context) error {
	userID := c.Get("user_id").(string) // Extracted seamlessly from Clerk context Middleware
	projects, err := h.service.Project.GetUserProjects(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, projects)
}

func (h *ProjectHandler) CreateProject(c echo.Context) error {
	userID := c.Get("user_id").(string)
	
	var input CreateProjectInput
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid format"})
	}

	result, err := h.service.Project.CreateProject(c.Request().Context(), input.Name, input.Description, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, result)
}
```

Register handler initialization in [apps/backend/internal/handler/handlers.go](apps/backend/internal/handler/handlers.go).

---

## Step 4: Register Routing Rules

Wire the Handler actions with router policies inside [apps/backend/internal/router/router.go](apps/backend/internal/router/router.go).

For authenticated routes, wrap with `middlewares.Auth.RequireAuth` to force verification:

```go
    // Mount authenticated routes under v1
    v1 := router.Group("/api/v1")
    
    // Pass JWT validation as mid-tier checks
    v1.GET("/projects", h.Project.GetProjects, middlewares.Auth.RequireAuth)
    v1.POST("/projects", h.Project.CreateProject, middlewares.Auth.RequireAuth)
```

Now, when starting your API via:
```bash
task run
```
You can access authenticated `/api/v1/projects` endpoints perfectly!

---

## Step 5: Implement UI and Client Interaction

Your frontend automatically knows about the new endpoint because `projects` was declared in the central TypeScript `apiContract`!

### A. Consume matching routes
In your view components inside [apps/frontend/src/](apps/frontend/src/), extract `useApiClient` and wrap it within a standard TanStack `useQuery`:

```tsx
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useApiClient } from "@/api/index";

export function UseProjects() {
  const api = useApiClient();
  const queryClient = useQueryClient();

  // Fully Typed return contract: compiler is aware it maps array of ZProject!
  const getProjectsQuery = useQuery({
    queryKey: ["projects"],
    queryFn: async () => {
      const resp = await api.getProjects();
      if (resp.status !== 200) throw new Error("Failed fetching resources");
      return resp.body;
    },
  });

  const createProjectMutation = useMutation({
    mutationFn: async (payload: { name: string; description?: string }) => {
      const resp = await api.createProject({ body: payload });
      if (resp.status !== 201) throw new Error("Creation went wrong");
      return resp.body;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });

  return {
    projects: getProjectsQuery.data || [],
    isLoading: getProjectsQuery.isLoading,
    createProject: createProjectMutation.mutate,
    isCreating: createProjectMutation.isPending,
  };
}
```

And just like that, the new feature is fully wired end-to-end with absolute, compile-safe types on both the backend models, api definitions, database state, and client-side react components.
