import WorkflowServicesPage from "~/pages/namespace/Explorer/Workflow/Services";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/n/$namespace/explorer/workflow/services/$"
)({
  component: WorkflowServicesPage,
});
