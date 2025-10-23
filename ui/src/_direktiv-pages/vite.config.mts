/// <reference types="vitest" />
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { viteSingleFile } from "vite-plugin-singlefile";
import viteTsconfigPaths from "vite-tsconfig-paths";

export default () =>
  defineConfig({
    root: "src/_direktiv-pages",
    server: {
      host: "0.0.0.0",
      port: 3001,
    },
    optimizeDeps: { esbuildOptions: { loader: { ".js": "jsx" } } },
    plugins: [react(), viteTsconfigPaths(), viteSingleFile()],
  });
