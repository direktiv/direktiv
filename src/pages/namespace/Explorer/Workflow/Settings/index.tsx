import Badge from "~/design/Badge";
import { Card } from "~/design/Card";
import { FC } from "react";
import { pages } from "~/util/router/pages";
import { useNodeContent } from "~/api/tree/query/node";
import { useWorkflowVariables } from "~/api/tree/query/variables";

const WorkflowSettingsPage: FC = () => {
  const { path } = pages.explorer.useParams();

  const { data } = useNodeContent({
    path,
  });

  const { data: variables } = useWorkflowVariables({
    path,
  });

  return (
    <div className="flex flex-col space-y-4 p-4">
      <h1>WorkflowSettingsPage</h1>
      <Card className="p-4">
        <Badge>{data?.revision?.hash.slice(0, 8)}</Badge>
      </Card>
      <Card className="p-4">
        <ul></ul>
        {variables?.variables.results.map((variable) => (
          <li key={variable.name}>{variable.name}</li>
        ))}
      </Card>
    </div>
  );
};

export default WorkflowSettingsPage;
