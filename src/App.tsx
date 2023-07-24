import "./App.css";
import "reactflow/dist/base.css";
import "./design/WorkflowDiagram/style.css";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useTheme, useThemeActions } from "~/util/store/theme";

import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { RouterProvider } from "react-router-dom";
import { Toaster } from "~/design/Toast";
import env from "./config/env/";
import { router } from "~/util/router";
import { useEffect } from "react";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
      networkMode: "always", // the default networkMode sometimes assumes that the client is offline
    },
    mutations: {
      retry: false,
      networkMode: "always", // the default networkMode sometimes assumes that the client is offline
    },
  },
});

const App = () => {
  const theme = useTheme();
  const { setSystemTheme } = useThemeActions();

  useEffect(() => {
    const systemTheme: typeof theme = window.matchMedia(
      "(prefers-color-scheme: dark)"
    ).matches
      ? "dark"
      : "light";
    setSystemTheme(systemTheme);

    // apply theme from local storage or system theme
    document
      .querySelector("html")
      ?.setAttribute("data-theme", theme ?? systemTheme);
  }, [setSystemTheme, theme]);

  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      {/* By default, React Query Devtools are only included in bundles when process.env.NODE_ENV === 'development', so you don't need to worry about excluding them during a production build. */}
      {env.VITE_RQ_DEV_TOOLS && <ReactQueryDevtools initialIsOpen={false} />}
      <Toaster />
    </QueryClientProvider>
  );
};

export default App;
