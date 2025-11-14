/// <reference types="vitest" />
import type { ViteDevServer } from "vite";
import { defineConfig } from "vite";
import { page } from "./src/examplePage";
import react from "@vitejs/plugin-react";
import { viteSingleFile } from "vite-plugin-singlefile";
import viteTsconfigPaths from "vite-tsconfig-paths";

function DirektivPagesMockPlugin() {
  return {
    name: "direktiv-pages-mock-plugin",
    configureServer(server: ViteDevServer) {
      server.middlewares.use("/page.json", (req, res) => {
        res.setHeader("Content-Type", "application/json");
        res.end(JSON.stringify(page));
      });
    },
  };
}

export default () =>
  defineConfig({
    root: "src/_direktiv-pages",
    server: {
      host: "0.0.0.0",
      port: 3001,
    },
    optimizeDeps: { esbuildOptions: { loader: { ".js": "jsx" } } },
    plugins: [
      react(),
      viteTsconfigPaths(),
      viteSingleFile(),
      DirektivPagesMockPlugin(),
    ],
  });
