import ServiceDetailsPage from "~/pages/namespace/Explorer/Workflow/Services/ServiceDetailsPage";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/n/$namespace/explorer/workflow/services/$service/$"
)({
  component: ServiceDetailsPage,
});
