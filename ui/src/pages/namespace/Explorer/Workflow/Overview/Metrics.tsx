import { Card } from "~/design/Card";
import Content from "./Content";
import { PieChart } from "lucide-react";
import RefreshButton from "~/design/RefreshButton";
import { useMetrics } from "~/api/metrics/query/metrics";
import { useTranslation } from "react-i18next";

const Metrics = ({ workflow }: { workflow: string }) => {
  const { t } = useTranslation();

  const { data, isFetching, isFetched, refetch } = useMetrics({
    path: workflow,
  });

  const MetricsRefetchButton = () => (
    <RefreshButton
      icon
      size="sm"
      variant="ghost"
      disabled={isFetching}
      onClick={() => {
        refetch();
      }}
    />
  );

  if (!data) return null;

  return (
    <Card className="flex flex-col">
      <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
        <PieChart className="h-5" />
        <h3 className="grow">
          {t("pages.explorer.tree.workflow.overview.metrics.header")}
        </h3>
        <MetricsRefetchButton />
      </div>
      <Content isFetched={isFetched} data={data} />
    </Card>
  );
};

export default Metrics;
