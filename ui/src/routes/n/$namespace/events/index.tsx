import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/events/")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/n/$namespace/events/"!</div>;
}
