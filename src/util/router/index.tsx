import Layout from "../../pages/namespace/Layout";
import { createBrowserRouter } from "react-router-dom";
import { pages } from "./pages";

export const router = createBrowserRouter([
  {
    path: "/:namespace",
    element: <Layout />,
    children: Object.values(pages).map((page) => page.route),
  },
  {
    path: "/",
    element: <Layout />,
    children: Object.values(pages).map((page) => page.route),
  },
]);
