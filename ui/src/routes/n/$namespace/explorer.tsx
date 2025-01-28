import { createFileRoute } from "@tanstack/react-router";

const ExplorerPage = () => (
  <div>
    <h1>Explorer Page</h1>
  </div>
);

export const Route = createFileRoute("/n/$namespace/explorer")({
  component: ExplorerPage,
});
