import "./App.css";
import "reactflow/dist/base.css";
import "./design/WorkflowDiagram/style.css";

import { useTheme, useThemeActions } from "~/util/store/theme";

import { AuthenticationProvider } from "./componentsNext/AuthenticationProvider";
import { OidcProvider } from "./componentsNext/OidcProvider";
import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { RouterProvider } from "react-router-dom";
import { Toaster } from "~/design/Toast";
import queryClient from "./util/queryClient";
import { router } from "~/util/router";
import { useEffect } from "react";

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
      <OidcProvider>
        <AuthenticationProvider>
          <RouterProvider router={router} />
        </AuthenticationProvider>
      </OidcProvider>
      {/* By default, React Query Devtools are only included in bundles when process.env.NODE_ENV === 'development', so you don't need to worry about excluding them during a production build. */}
      {!!process.env.VITE?.VITE_RQ_DEV_TOOLS && (
        <ReactQueryDevtools initialIsOpen={false} />
      )}
      <Toaster />
    </QueryClientProvider>
  );
};

export default App;
