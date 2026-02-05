import { Orientation } from "./types";
import { ReactFlowProvider } from "reactflow";
import { WorkflowStatesSchemaType } from "~/api/instances/schema";
import { ZoomPanDiagram } from "./ZoomPanDiagram";
import { createElements } from "./utils";

/**
 * Renders a diagram of a workflow and optionally its current state position during a instance.
 * * Props
 *   * states: JSON describing the states of the workflow.
 *   * instanceStatus: Status of current instance. This is used to display if flow is complete with animated connections.
 *   * disabled: Disables diagram zoom-in
 */
type WorkflowDiagramProps = {
  states: WorkflowStatesSchemaType;
  orientation?: Orientation;
  instanceStatus?: "pending" | "complete" | "failed";
  disabled?: boolean;
};

export default function WorkflowDiagram(props: WorkflowDiagramProps) {
  const {
    states,
    instanceStatus = "pending",
    disabled = false,
    orientation = "horizontal",
  } = props;

  const flowElements = createElements(states, instanceStatus, orientation);

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
