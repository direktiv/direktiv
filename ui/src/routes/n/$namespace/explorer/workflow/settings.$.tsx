import WorkflowSettingsPage from "~/pages/namespace/Explorer/Workflow/Settings";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/n/$namespace/explorer/workflow/settings/$"
)({
  component: WorkflowSettingsPage,
});
