import ErrorPage from "./ErrorPage";
import NamespaceLayout from "~/pages/namespace/Layout";
import OnboardingPage from "~/pages/OnboardingPage";
import { createBrowserRouter } from "react-router-dom";
import { pages } from "./pages";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <OnboardingPage />,
    errorElement: <ErrorPage />,
  },
  {
    path: "/n/:namespace",
    element: <NamespaceLayout />,
    children: Object.values(pages).map((page) => page.route),
    errorElement: <ErrorPage />,
  },
]);
