import Badge from "../../../../../design/Badge";
import { Card } from "../../../../../design/Card";
import { FC } from "react";
import { pages } from "../../../../../util/router/pages";
import { useNodeContent } from "../../../../../api/tree/query/get";

const WorkflowOverviewPage: FC = () => {
  const { path } = pages.explorer.useParams();
  const { data } = useNodeContent({
    path,
  });

  return (
    <div className="flex flex-col space-y-4 p-4">
      <h1>WorkflowOverviewPage</h1>
      <Card className="flex flex-col space-y-4 p-4">
        <div>
          <Badge>{data?.revision?.hash.slice(0, 8)}</Badge>
        </div>
        <Card className="p-4">
          {data?.revision?.source && atob(data?.revision?.source)}
        </Card>
      </Card>
    </div>
  );
};

export default WorkflowOverviewPage;
