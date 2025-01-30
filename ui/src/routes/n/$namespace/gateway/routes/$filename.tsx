import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/gateway/routes/$filename")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/n/$namespace/gateway/routes/$filename"!</div>;
}
