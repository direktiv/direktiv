import Badge from "../../../../../design/Badge";
import { Card } from "../../../../../design/Card";
import { FC } from "react";
import { pages } from "../../../../../util/router/pages";
import { useNodeContent } from "../../../../../api/tree/query/get";

const WorkflowRevisionsPage: FC = () => {
  const { path } = pages.explorer.useParams();
  const { data } = useNodeContent({
    path,
  });

  return (
    <div className="flex flex-col space-y-4 p-4">
      <h1>WorkflowRevisionsPage</h1>
      <Card className="p-4">
        <Badge>{data?.revision?.hash.slice(0, 8)}</Badge>
      </Card>
    </div>
  );
};

export default WorkflowRevisionsPage;
