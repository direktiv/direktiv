import "./App.css";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { RouterProvider } from "react-router-dom";
import { router } from "./util/router";
import { useEffect } from "react";
import { useTheme } from "./util/store/theme";

const queryClient = new QueryClient();

const App = () => {
  const theme = useTheme();
  useEffect(() => {
    let applyTheme = window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";

    if (theme) applyTheme = theme;
    document.querySelector("html")?.setAttribute("data-theme", applyTheme);
  }, [theme]);

  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
};

export default App;
