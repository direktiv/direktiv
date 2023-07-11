import "./style.css";
import "reactflow/dist/base.css";

import { generateElements, getLayoutedElements } from "./utils";
import { useMemo, useState } from "react";

import Alert from "../Alert";
import { IWorkflow } from "./types";
import { ReactFlowProvider } from "reactflow";
import YAML from "js-yaml";
import { ZoomPanDiagram } from "./ZoomPanDiagram";

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
  instanceStatus?: "pending" | "complete" | "failed";
  disabled?: boolean;
};

export default function WorkflowDiagram(props: WorkflowDiagramProps) {
  const {
    workflow,
    flow = [],
    instanceStatus = "pending",
    disabled = false,
  } = props;

  const [invalidWorkflow, setInvalidWorkflow] = useState<string | null>(null);

  const parsedWorkflow = useMemo(() => {
    if (!workflow) {
      setInvalidWorkflow(null);
      return null;
    }
    try {
      const workflowObj = YAML.load(workflow) as IWorkflow;
      setInvalidWorkflow(null);
      return workflowObj;
    } catch (error: unknown) {
      setInvalidWorkflow(error?.toString() ?? "Unknown error");
      return null;
    }
  }, [workflow]);

  if (invalidWorkflow) return <Alert variant="error">{invalidWorkflow}</Alert>;
  if (parsedWorkflow === null) return null;

  const flowElements = generateElements(
    getLayoutedElements,
    parsedWorkflow,
    flow,
    instanceStatus
  );

  if (flowElements.length === 0) return null;

  return (
    <ReactFlowProvider>
      <ZoomPanDiagram disabled={disabled} elements={flowElements} />
    </ReactFlowProvider>
  );
}
