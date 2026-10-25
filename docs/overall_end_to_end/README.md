# End-to-End Architecture & Guide

Welcome to the **Go-tasker** end-to-end documentation! Since you have just cloned or forked this repository and are a beginner to this project and tech stack, this guide is crafted specifically for you.

This monorepo is a production-ready template designed with clean architecture, enterprise-level performance, solid type safety, background queues, and complete system observability. It is much more than a simple prototype; it implements real-world patterns used by experienced software engineering teams.

---

## Guide Index

To help you learn systematically, we have broken down this documentation into three core chapters:

1. **[docs/overall_end_to_end/architecture_and_tech_stack.md](docs/overall_end_to_end/architecture_and_tech_stack.md)**
   * Learn *what* technologies are used, *why* they were chosen, *how* they connect, and what their specific "job" is in the tasker.
   * Understand the flow of requests from the user's browser, through authentication, routing, services, background jobs, database, and observability agents.

2. **[docs/overall_end_to_end/how_it_differs_from_todo_app.md](docs/overall_end_to_end/how_it_differs_from_todo_app.md)**
   * A detailed contrast showing how a production-ready architecture differs fundamentally from a traditional, naive frontend/backend tutorial "TODO app".
   * Covers database connection reliability, background task offloading, contract-first API safety, structured logging vs. stdout `println`, and automated integration testing.

3. **[docs/overall_end_to_end/how_to_add_features.md](docs/overall_end_to_end/how_to_add_features.md)**
   * A step-by-step developer playbook detailing how to build and integrate any new feature end-to-end.
   * Walks through schema definitions, contract composition, running codegen, writing Go database migrations, exposing backend services/handlers, routing, and wiring React query hooks on the frontend.

---

## Core Directories & Navigating the Monorepo

The codebase is organized as a monorepo managed with **Turborepo** and **Bun Workspaces**. Here is a high-level breakdown of the folders in this repository:

* **[apps/backend](apps/backend)**:
  * Written in Go 1.24+ using the Echo web framework.
  * Employs clean architecture layers (Handlers → Services → Repositories → Models).
  * Manages database migrations with Tern, jobs with Asynq, and analytics with New Relic.

* **[apps/frontend](apps/frontend)**:
  * Built with React 19, Vite 7, and TailwindCSS v4.
  * Uses TanStack React Query for data fetching, Axios for the underlying HTTP requests, and Clerk for user registration and authentication.

* **[packages/openapi](packages/openapi)**:
  * The Single Source of Truth (SSOT) of the API.
  * Expresses endpoints as TypeScript definitions using ts-rest and packages them as an NPM workspace dependency.

* **[packages/zod](packages/zod)**:
  * Houses shared Zod validation schemas. Used by ts-rest to validate request parameters, bodies, and responses before compilation and at runtime.

* **[packages/emails](packages/emails)**:
  * TypeScript templates using React Email for beautiful, standardized emails that are rendered and delivered transactional-style.

---

## Getting Started Strategy for Beginners

If you are new to Go, React, SQL, or Docker:
1. Start by executing the backend locally and querying the `/status` endpoint to make sure you can see a `"healthy"` response.
2. Read the **[Chapter 1: Tech Stack & Connections](docs/overall_end_to_end/architecture_and_tech_stack.md)** to grasp how components like Redis and PostgreSQL work together.
3. Keep the **[Chapter 3: How-To Playbook](docs/overall_end_to_end/how_to_add_features.md)** open as you build your very first endpoint (e.g., adding an API that lists resources or logs statistics).

You've got a fantastic, modern codebase in front of you. Let's dive in!
