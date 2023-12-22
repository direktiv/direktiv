import ErrorPage from "./ErrorPage";
import NamespaceLayout from "~/pages/namespace/Layout";
import OnboardingPage from "~/pages/OnboardingPage";
import { createBrowserRouter } from "react-router-dom";
import env from "~/config/env";
import { pages } from "./pages";

export const router = createBrowserRouter(
  [
    {
      path: "/",
      element: <OnboardingPage />,
      errorElement: <ErrorPage />,
    },
    {
      path: "/:namespace",
      element: <NamespaceLayout />,
      children: Object.values(pages).map((page) => page.route),
      errorElement: <ErrorPage />,
    },
  ],
  {
    basename: process.env.VITE?.VITE_BASE ?? undefined,
  }
);
