import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import viteTsconfigPaths from "vite-tsconfig-paths";
import svgrPlugin from "vite-plugin-svgr";
import path from "path";
import fs from "fs";

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
  const { VITE_DEV_API_DOMAIN } = loadEnv(mode, process.cwd());

  if (!VITE_DEV_API_DOMAIN) {
    console.warn("VITE_DEV_API_DOMAIN is not set, no API proxy will be used");
  }

  return defineConfig({
    server: {
      host: "0.0.0.0",
      port: 3000,
      proxy: VITE_DEV_API_DOMAIN
        ? {
            "/api": {
              target: VITE_DEV_API_DOMAIN,
            },
          }
        : {},
    },
    optimizeDeps: {
      esbuildOptions: {
        loader: {
          ".js": "jsx",
        },
      },
    },
    plugins: [react(), viteTsconfigPaths(), svgrPlugin(), reactVirtualized()],
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
      ],
    },
  });
};
