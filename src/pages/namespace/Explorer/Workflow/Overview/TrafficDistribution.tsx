import { Card } from "~/design/Card";
import { CategoryBar } from "@tremor/react";
import { Network } from "lucide-react";
import { NoResult } from "~/design/Table";
import { useRouter } from "~/api/tree/query/router";
import { useTranslation } from "react-i18next";

const TrafficDistribution = ({ workflow }: { workflow: string }) => {
  const { t } = useTranslation();

  const { data: routerData } = useRouter({ path: workflow });

  const routes = routerData?.routes;

  return (
    <Card className="flex flex-col">
      <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
        <Network className="h-5" />
        <h3 className="grow">
          {t(
            "pages.explorer.tree.workflow.overview.trafficDistribution.header"
          )}
        </h3>
      </div>
      {routes && routes[0] && routes[1] ? (
        <div className="flex h-full flex-col justify-center p-5 pt-1">
          <CategoryBar
            values={[routes[0].weight, routes[1].weight]}
            colors={["indigo", "gray"]}
            markerValue={routes[0].weight}
            className="mt-3"
          />
          <div className="flex flex-row justify-between">
            <div>{routes[0].ref.slice(0, 8)}</div>
            <div>{routes[1].ref.slice(0, 8)}</div>
          </div>
        </div>
      ) : (
        <NoResult>
          {t(
            "pages.explorer.tree.workflow.overview.trafficDistribution.noResult"
          )}
        </NoResult>
      )}
    </Card>
  );
};

export default TrafficDistribution;
