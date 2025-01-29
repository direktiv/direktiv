import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/jq")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/n/$namespace/jqplayground"!</div>;
}
