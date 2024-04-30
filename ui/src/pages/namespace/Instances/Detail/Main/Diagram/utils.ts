import { ComponentProps } from "react";
import { InstanceSchemaType } from "~/api/instances/schema";
import WorkflowDiagram from "~/design/WorkflowDiagram";

type DiagramStatus = ComponentProps<typeof WorkflowDiagram>["instanceStatus"];

export const instanceStatusToDiagramStatus = (
  status: InstanceSchemaType["status"]
): DiagramStatus => {
  switch (status) {
    case "failed":
      return "failed";
    case "complete":
      return "complete";
    case "crashed":
      return "failed";
    case "pending":
    default:
      return undefined;
  }
};
