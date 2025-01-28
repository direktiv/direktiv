import { Outlet, createRootRouteWithContext } from "@tanstack/react-router";

import { QueryClient } from "@tanstack/react-query";

const RootComponent = () => <Outlet />;

export const Route = createRootRouteWithContext<{
  queryClient: QueryClient;
  apiKey: string | undefined;
}>()({
  component: RootComponent,
});
