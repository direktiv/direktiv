import { RouterProvider, createRouter } from "@tanstack/react-router";

import ErrorPage from "~/util/router/ErrorPage";
import NotFoundPage from "~/util/router/NotFoundPage";
import queryClient from "~/util/queryClient";
import { routeTree } from "~/routeTree.gen";

const router = createRouter({
  routeTree,
  context: { queryClient, apiKey: undefined },
  defaultErrorComponent: ErrorPage,
  defaultNotFoundComponent: NotFoundPage,
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

export const Router = () => <RouterProvider router={router} />;
