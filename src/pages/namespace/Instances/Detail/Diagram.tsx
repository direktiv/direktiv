import { FC } from "react";
import WorkflowDiagram from "~/design/WorkflowDiagram";
import { useNodeContent } from "~/api/tree/query/node";

const Diagram: FC<{ workflowPath: string }> = ({ workflowPath }) => {
  const { data } = useNodeContent({ path: workflowPath });
  if (!data) return null;

  const workflowData = atob(data.revision?.source ?? "");
  return <WorkflowDiagram workflow={workflowData} />;
};

export default Diagram;
