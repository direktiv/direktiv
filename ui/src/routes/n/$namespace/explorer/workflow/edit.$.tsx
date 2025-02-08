import WorkflowEditPage from "~/pages/namespace/Explorer/Workflow/Edit";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer/workflow/edit/$")({
  component: WorkflowEditPage,
});
