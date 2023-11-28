/// <reference types="vitest" />
import { defineConfig, loadEnv } from "vite";

import { envVariablesSchema } from "./src/config/env/schema";
import pluginRewriteAll from "vite-plugin-rewrite-all";
import react from "@vitejs/plugin-react";
import svgrPlugin from "vite-plugin-svgr";
import viteTsconfigPaths from "vite-tsconfig-paths";

export default ({ mode }) => {
  const env = loadEnv(mode, process.cwd());

  const parsedEnv = envVariablesSchema.parse(env);

  const { VITE_DEV_API_DOMAIN: apiDomain } = parsedEnv;

  const baseconfig = env.VITE_BASE ? { base: env.VITE_BASE } : {};

  if (!apiDomain) {
    console.warn("VITE_DEV_API_DOMAIN is not set, no API proxy will be used");
  }

  return defineConfig({
    server: {
      host: "0.0.0.0",
      port: 3000,
      proxy: apiDomain
        ? {
            "/api": {
              target: apiDomain,
              secure: false,
            },
            "/oidc": {
              target: apiDomain,
              secure: false,
            },
          }
        : {},
    },
    build: {
      commonjsOptions: {
        // https://github.com/vitejs/vite/issues/2139#issuecomment-1405624744
        defaultIsModuleExports(id) {
          try {
            // eslint-disable-next-line @typescript-eslint/no-var-requires
            const module = require(id);
            if (module?.default) {
              return false;
            }
            return "auto";
          } catch (error) {
            return "auto";
          }
        },
      },
    },
    optimizeDeps: {
      esbuildOptions: {
        loader: {
          ".js": "jsx",
        },
      },
    },
    plugins: [react(), viteTsconfigPaths(), svgrPlugin(), pluginRewriteAll()],
    test: {
      globals: true,
      environment: "jsdom",
      exclude: [
        "**/node_modules/**",
        "**/dist/**",
        "**/cypress/**",
        "**/.{idea,git,cache,output,temp}/**",
        "**/{karma,rollup,webpack,vite,vitest,jest,ava,babel,nyc,cypress,tsup,build}.config.*",
        "e2e/**", // playwright tests, vitest throws errors when parsing them.
      ],
    },
    ...baseconfig,
  });
};
