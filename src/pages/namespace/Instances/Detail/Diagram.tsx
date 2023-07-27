import { FC } from "react";
import WorkflowDiagram from "~/design/WorkflowDiagram";
import { useNodeContent } from "~/api/tree/query/node";

const Diagram: FC<{ workflowPath: string; flow: string[] }> = ({
  workflowPath,
  flow,
}) => {
  const { data } = useNodeContent({ path: workflowPath });
  if (!data) return null;

  const workflowData = atob(data.revision?.source ?? "");
  return (
    <WorkflowDiagram
      workflow={workflowData}
      flow={flow}
      orientation="vertical"
    />
  );
};

export default Diagram;
