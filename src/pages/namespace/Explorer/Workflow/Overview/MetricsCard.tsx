import { Card } from "~/design/Card";
import { DonutChart } from "@tremor/react";
import { PieChart } from "lucide-react";
import { useMetrics } from "~/api/tree/query/metrics";
import { useTranslation } from "react-i18next";

const MetricsCard = ({ path }: { path: string }) => {
  const { t } = useTranslation();
  const { data: successData } = useMetrics({ path, type: "successful" });
  const { data: failedData } = useMetrics({ path, type: "failed" });

  const successful = successData?.results[0]?.value[1];
  const failed = failedData?.results[0]?.value[1];
  const total = successful + failed;

  const metrics = [
    {
      name: "failed",
      count: failed,
    },
    {
      name: "successful",
      count: successful,
    },
  ];

  const percentages = {
    successful: (successful / total) * 100,
    failed: (failed / total) * 100,
  };

  const valueFormatter = (number: number) => number.toString();

  return (
    <Card className="flex flex-col">
      <div className="flex items-center gap-x-2 border-b border-gray-5 p-5 font-medium dark:border-gray-dark-5">
        <PieChart className="h-5" />
        <h3 className="grow">
          {t("pages.explorer.tree.workflow.overview.metrics.header")}
        </h3>
      </div>
      <DonutChart
        noDataText="TBD i18n key"
        showAnimation={false}
        showLabel={false}
        className="mt-6"
        data={metrics}
        category="count"
        index="name"
        valueFormatter={valueFormatter}
        colors={["red", "green"]}
      />
      <div className="mb-5 flex justify-evenly">
        <div>
          {t("pages.explorer.tree.workflow.overview.metrics.successful", {
            percentage: percentages.successful.toFixed(0),
          })}
        </div>
        <div>
          {t("pages.explorer.tree.workflow.overview.metrics.failed", {
            percentage: percentages.failed.toFixed(0),
          })}
        </div>
      </div>
    </Card>
  );
};

export default MetricsCard;
