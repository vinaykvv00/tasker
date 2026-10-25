# Frontend Architecture & Reference Manual

Welcome to the frontend reference guide! The client-side application in **Go-tasker** is located inside the [apps/frontend](apps/frontend) directory. Built on modern web standards, it prioritizes performance, strict developer type safety, and real-time form validation.

---

## 1. Directory Structure

The react client is organized with the following structure:

* **[apps/frontend/src/api/index.ts](apps/frontend/src/api/index.ts)**:
  * Initializes the `@ts-rest/core` API client using contract specifications from our shared schema package.
  * Intercepts outgoing client inquiries to automatically attach Clerk Bearer authentication headers.
  * Incorporates automatic connection state retries on standard HTTP 401 Unauthorized status changes.

* **[apps/frontend/src/config/env.ts](apps/frontend/src/config/env.ts)**:
  * Validates environment variables (e.g. `VITE_API_URL` and `VITE_CLERK_PUBLISHABLE_KEY`) on startup.

* **[apps/frontend/src/App.tsx](apps/frontend/src/App.tsx)**:
  * The main shell of the frontend presentation rendering, handling routing and custom view pages.

* **[apps/frontend/src/main.tsx](apps/frontend/src/main.tsx)**:
  * Boostraps React 19, attaches global styling rules, and wraps the DOM with React-query and Clerk authentication providers.

* **[apps/frontend/src/index.css](apps/frontend/src/index.css)**:
  * Imports the TailwindCSS utility layout definitions.

---

## 2. API Contract & Shared Workspace Packages

Rather than writing fetch pathways manually, the frontend accesses external APIs through shared packages defined directly inside the workspace:

* **[packages/openapi](packages/openapi)**:
  * Exposes TypeScript contracts mapping route URLs, types, methods, request parameters, bodies, and responses.

* **[packages/zod](packages/zod)**:
  * Exposes Zod structures used to parse frontend data, handle runtime validation checks, and power form schemas.

When you modify schemas inside these packages and run `bun build` at the root directory, TypeScript compiles the files. Consequently, all endpoints accessed via `useApiClient()` are immediately aware of the code changes, validating your code at build time.

---

## 3. Developer Workflow Commands

You can trigger tasks in the frontend using Bun directly:

* **`bun dev`**: Start the Vite web server locally. Automatically waits for packages configuration changes.
* **`bun build`**: Compiles TS types, parses env parameters, and compiles static SPA bundles into the `dist/` directory.
* **`bun lint`**: Examines code cleanliness using ESLint rules.
* **`bun format:fix`**: Automatically formats markup, styles, and scripts with Prettier formatting.
