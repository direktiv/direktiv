import { Card } from "~/design/Card";
import { NoResult } from "~/design/Table";
import { PieChart } from "lucide-react";
import RefreshButton from "~/design/RefreshButton";
import SuccessFailure from "./SuccessFailure";
import { useMetrics } from "~/api/tree/query/metrics";
import { useTranslation } from "react-i18next";

const Metrics = ({ workflow }: { workflow: string }) => {
  const { t } = useTranslation();

  const {
    data: successData,
    isFetching: isFetchingSuccessful,
    refetch: refetchSuccessful,
  } = useMetrics({
    path: workflow,
    type: "successful",
  });
  const {
    data: failedData,
    isFetching: isFetchingFailed,
    refetch: refetchFailed,
  } = useMetrics({
    path: workflow,
    type: "failed",
  });

  const successful = Number(successData?.results[0]?.value[1]);
  const failed = Number(failedData?.results[0]?.value[1]);
  const metrics =
    successful || failed
      ? {
          successful: successful || 0,
          failed: failed || 0,
        }
      : undefined;

  const isFetchingMetrics = isFetchingFailed || isFetchingSuccessful;

  const refetchMetrics = () => {
    refetchSuccessful();
    refetchFailed();
  };

  const MetricsRefetchButton = () => (
    <RefreshButton
      icon
      size="sm"
      variant="ghost"
      disabled={isFetchingMetrics}
      onClick={() => {
        refetchMetrics();
      }}
    />
  );

  return (
    <Card className="flex flex-col">
      <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
        <PieChart className="h-5" />
        <h3 className="grow">
          {t("pages.explorer.tree.workflow.overview.metrics.header")}
        </h3>
        <MetricsRefetchButton />
      </div>
      {metrics ? (
        <SuccessFailure data={metrics} />
      ) : (
        <NoResult icon={PieChart}>
          {t("pages.explorer.tree.workflow.overview.metrics.noResult")}
        </NoResult>
      )}
    </Card>
  );
};

export default Metrics;
