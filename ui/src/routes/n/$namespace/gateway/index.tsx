import { Outlet, createFileRoute } from "@tanstack/react-router";

const Routes = () => <Outlet />;

export const Route = createFileRoute("/n/$namespace/gateway/")({
  component: Routes,
});
