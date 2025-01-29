import { createFileRoute } from "@tanstack/react-router";

const ServiceDetailPage = () => (
  <div>
    <h1>Service detail page</h1>
  </div>
);

export const Route = createFileRoute("/n/$namespace/services/$service")({
  component: ServiceDetailPage,
});
