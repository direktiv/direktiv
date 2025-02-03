import WorkflowEditor from "~/pages/namespace/Explorer/Workflow/Edit/WorkflowEditor";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer/workflow/edit/$")({
  component: WorkflowEditor,
});
