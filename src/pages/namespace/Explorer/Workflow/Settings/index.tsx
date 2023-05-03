import Badge from "../../../../../design/Badge";
import { Card } from "../../../../../design/Card";
import { FC } from "react";
import { pages } from "../../../../../util/router/pages";
import { useListDirectory } from "../../../../../api/tree/query/get";

const WorkflowSettingsPage: FC = () => {
  const { path } = pages.explorer.useParams();
  const { data } = useListDirectory({
    path,
  });

  return (
    <div className="flex flex-col space-y-4 p-4">
      <h1>WorkflowSettingsPage</h1>
      <Card className="p-4">
        <Badge>{data?.revision?.hash.slice(0, 8)}</Badge>
      </Card>
    </div>
  );
};

export default WorkflowSettingsPage;
