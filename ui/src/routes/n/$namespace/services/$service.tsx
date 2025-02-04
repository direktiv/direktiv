import ServiceDetailPage from "~/pages/namespace/Services/Detail";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/services/$service")({
  component: ServiceDetailPage,
});
