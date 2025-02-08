import { RouterProvider, createRouter } from "@tanstack/react-router";

import queryClient from "~/util/queryClient";
import { routeTree } from "~/routeTree.gen";

const router = createRouter({
  routeTree,
  context: { queryClient, apiKey: undefined },
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

export const Router = () => <RouterProvider router={router} />;
