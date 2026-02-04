import Alert from "../Alert";
import { InstanceFlowResponse } from "~/api/instances/schema";
import { Orientation } from "./types";
import { ReactFlowProvider } from "reactflow";
import { ZoomPanDiagram } from "./ZoomPanDiagram";
import { createElements } from "./utils";
import { useState } from "react";

/**
 * Renders a diagram of a workflow and optionally its current state position during a instance.
 * * Props
 *   * workflow: JSON of workflow.
 *   * instanceStatus: Status of current instance. This is used to display if flow is complete with animated connections.
 *   * disabled: Disables diagram zoom-in
 */
type WorkflowDiagramProps = {
  workflow: InstanceFlowResponse;
  orientation?: Orientation;
  instanceStatus?: "pending" | "complete" | "failed";
  disabled?: boolean;
};

export default function WorkflowDiagram(props: WorkflowDiagramProps) {
  const {
    workflow,
    instanceStatus = "pending",
    disabled = false,
    orientation = "horizontal",
  } = props;

  const [invalidWorkflow] = useState<string | null>(null);

  if (invalidWorkflow)
    return (
      <Alert className="flex" variant="error">
        {invalidWorkflow}
      </Alert>
    );

  const flowElements = createElements(workflow, instanceStatus, orientation);

  return (
    <ReactFlowProvider>
      <ZoomPanDiagram
        disabled={disabled}
        elements={flowElements}
        orientation={orientation}
      />
    </ReactFlowProvider>
  );
}
