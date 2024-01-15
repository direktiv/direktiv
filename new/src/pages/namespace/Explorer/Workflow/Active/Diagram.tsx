import { Card } from "~/design/Card";
import { FC } from "react";
import { LayoutsType } from "~/util/store/editor";
import WorkflowDiagram from "~/design/WorkflowDiagram";

type DiagramProps = {
  layout: LayoutsType;
  workflowData: string;
};

export const Diagram: FC<DiagramProps> = ({ layout, workflowData }) => (
  <Card className="flex grow" data-testid="workflow-diagram">
    <WorkflowDiagram
      workflow={workflowData}
      orientation={layout === "splitVertically" ? "vertical" : "horizontal"}
    />
  </Card>
);
