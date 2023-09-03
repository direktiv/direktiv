import Badge from "~/design/Badge";
import { Card } from "~/design/Card";
import { FC } from "react";
import { pages } from "~/util/router/pages";
import { useMetrics } from "~/api/tree/query/metrics";
import { useNodeContent } from "~/api/tree/query/node";
import { useRouter } from "~/api/tree/query/router";

const ActiveWorkflowPage: FC = () => {
  const { path } = pages.explorer.useParams();
  const { data } = useNodeContent({
    path,
  });
  const { data: routerData } = useRouter({ path });
  const { data: successData } = useMetrics({ path, type: "successful" });
  const { data: failedData } = useMetrics({ path, type: "failed" });

  const routes = routerData?.routes;

  return (
    <div className="flex flex-col space-y-4 p-4">
      <Card className="p-4">
        <Badge>{data?.revision?.hash.slice(0, 8)}</Badge>
      </Card>
      <Card className="p-4">
        <ul>
          <li>
            Traffic distribution:{" "}
            {routes && routes[0] && routes[1]
              ? `${routes[0].ref}: ${routes[0].weight} - ${routes[1].ref}: ${routes[1].weight}`
              : "not configured"}
          </li>
          <li>
            Success / failure rate: {successData?.results[0]?.value[1]}/
            {failedData?.results[0]?.value[1]}
          </li>
        </ul>
      </Card>
    </div>
  );
};

export default ActiveWorkflowPage;
