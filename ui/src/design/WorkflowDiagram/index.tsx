import { Orientation, Workflow } from "./types";
import { generateElements, getLayoutedElements } from "./utils";
import { useMemo, useState } from "react";

import Alert from "../Alert";
import { ReactFlowProvider } from "reactflow";
import { ZoomPanDiagram } from "./ZoomPanDiagram";
import { parse } from "yaml";

/**
 * Renders a diagram of a workflow and optionally its current state position during a instance.
 * * Props
 *   * workflow: YAML string of workflow.
 *   * flow: Array of executed states in an instance. Example - ['noopA', 'noopB']
 *   * instanceStatus: Status of current instance. This is used to display if flow is complete with animated connections.
 *   * disabled: Disables diagram zoom-in
 */
type WorkflowDiagramProps = {
  workflow: string;
  flow?: string[];
  orientation?: Orientation;
  instanceStatus?: "pending" | "complete" | "failed";
  disabled?: boolean;
};

export default function WorkflowDiagram(props: WorkflowDiagramProps) {
  const {
    workflow,
    flow = [],
    instanceStatus = "pending",
    disabled = false,
    orientation = "horizontal",
  } = props;

  const [invalidWorkflow, setInvalidWorkflow] = useState<string | null>(null);

  const parsedWorkflow = useMemo(() => {
    if (!workflow) {
      setInvalidWorkflow(null);
      return null;
    }
    try {
      const workflowObj = parse(workflow) as Workflow;
      setInvalidWorkflow(null);
      return workflowObj;
    } catch (error: unknown) {
      setInvalidWorkflow(error?.toString() ?? "Unknown error");
      return null;
    }
  }, [workflow]);

  if (invalidWorkflow)
    return (
      <Alert className="flex" variant="error">
        {invalidWorkflow}
      </Alert>
    );

  const flowElements = generateElements(
    getLayoutedElements,
    parsedWorkflow,
    flow,
    instanceStatus,
    orientation
  );

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
