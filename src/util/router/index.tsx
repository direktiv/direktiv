import Layout from "../../pages/layout";
import { createBrowserRouter } from "react-router-dom";
import { pages } from "./pages";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
    children: Object.values(pages).map((page) => page.route),
  },
]);
