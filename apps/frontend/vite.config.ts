import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import path from "path";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  define: {
    // only expose the VITE_-prefixed vars the app needs, not the full host/CI environment
    "process.env": {
      VITE_CLERK_PUBLISHABLE_KEY: process.env.VITE_CLERK_PUBLISHABLE_KEY,
      VITE_API_URL: process.env.VITE_API_URL,
      VITE_ENV: process.env.VITE_ENV,
    },
  },
  server: {
    port: 3000,
  },
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
      "@tasker/openapi": path.resolve(__dirname, "../../packages/openapi/src"),
      "@tasker/zod": path.resolve(__dirname, "../../packages/zod/src"),
    },
  },
});
