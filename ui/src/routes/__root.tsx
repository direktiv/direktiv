import { Outlet, createRootRouteWithContext } from "@tanstack/react-router";

import { QueryClient } from "@tanstack/react-query";

const RootComponent = () => (
  <div className="flex h-screen">
    <Outlet />
  </div>
);

export const Route = createRootRouteWithContext<{
  queryClient: QueryClient;
  apiKey: string | undefined;
}>()({
  component: RootComponent,
});
