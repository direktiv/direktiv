import "./App.css";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useTheme, useThemeActions } from "./util/store/theme";

import { RouterProvider } from "react-router-dom";
import { Toaster } from "./componentsNext/Toast";
import { router } from "./util/router";
import { useEffect } from "react";

const queryClient = new QueryClient();

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
      <Toaster />
    </QueryClientProvider>
  );
};

export default App;
