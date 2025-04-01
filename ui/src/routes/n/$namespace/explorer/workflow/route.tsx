import WorkflowLayout from "~/pages/namespace/Explorer/Workflow";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer/workflow")({
  component: WorkflowLayout,
});
