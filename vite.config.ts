/// <reference types="vitest" />
import { defineConfig, loadEnv } from "vite";

import { envVariablesSchema } from "./src/config/env/schema";
import fs from "fs";
import path from "path";
import pluginRewriteAll from "vite-plugin-rewrite-all";
import react from "@vitejs/plugin-react";
import svgrPlugin from "vite-plugin-svgr";
import viteTsconfigPaths from "vite-tsconfig-paths";

//  fix https://github.com/uber/baseweb/issues/4129
const WRONG_CODE = `import { bpfrpt_proptype_WindowScroller } from "../WindowScroller.js";`;
export function reactVirtualized() {
  return {
    name: "my:react-virtualized",
    configResolved() {
      const file = require
        .resolve("react-virtualized")
        .replace(
          path.join("dist", "commonjs", "index.js"),
          path.join("dist", "es", "WindowScroller", "utils", "onScroll.js")
        );
      const code = fs.readFileSync(file, "utf-8");
      const modified = code.replace(WRONG_CODE, "");
      fs.writeFileSync(file, modified);
    },
  };
}

export default ({ mode }) => {
  const env = loadEnv(mode, process.cwd());

  const parsedEnv = envVariablesSchema.parse(env);

  const { VITE_DEV_API_DOMAIN: apiDomain } = parsedEnv;

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
    plugins: [
      react(),
      viteTsconfigPaths(),
      svgrPlugin(),
      reactVirtualized(),
      pluginRewriteAll(),
    ],
    test: {
      globals: true,
      environment: "jsdom",
      exclude: [
        "**/node_modules/**",
        "**/dist/**",
        "**/cypress/**",
        "**/.{idea,git,cache,output,temp}/**",
        "**/{karma,rollup,webpack,vite,vitest,jest,ava,babel,nyc,cypress,tsup,build}.config.*",
        // all above this line are the default
        "src/hooks/**/*", // 🚧 search for TODO_HOOKS_TESTS to find all places that needs some action 🚧
        "e2e/**", // playwright tests, vitest throws errors when parsing them.
      ],
    },
  });
};
