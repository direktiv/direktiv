import {
  Navigate,
  RouterProvider as RouterProviderReactRouterDom,
  createBrowserRouter,
} from "react-router-dom";

import ErrorPage from "./ErrorPage";
import NamespaceLayout from "~/pages/namespace/Layout";
import OnboardingPage from "~/pages/OnboardingPage";
import { usePages } from "./pages";

export const RouterProvider = () => {
  const pages = usePages();
  const router = createBrowserRouter([
    {
      path: "/",
      element: <OnboardingPage />,
      errorElement: <ErrorPage />,
    },
    {
      // this is the OIDC redirect route
      path: "/callback",
      element: <Navigate to="/" replace />,
    },
    {
      path: "/n/:namespace",
      element: <NamespaceLayout />,
      children: Object.values(pages).map((page) => page.route),
      errorElement: <ErrorPage />,
    },
  ]);

  return <RouterProviderReactRouterDom router={router} />;
};
