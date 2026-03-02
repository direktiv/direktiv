import WorkflowServicesListPage from "~/pages/namespace/Explorer/Workflow/Services/ServicesListPage";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/n/$namespace/explorer/workflow/services/list/$"
)({
  component: WorkflowServicesListPage,
});
