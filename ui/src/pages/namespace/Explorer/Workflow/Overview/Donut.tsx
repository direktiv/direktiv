import { DonutChart } from "@tremor/react";
import { MetricsObjectSchemaType } from "~/api/metrics/schema";
import { useTranslation } from "react-i18next";

const Donut = ({ data }: { data: MetricsObjectSchemaType }) => {
  const { t } = useTranslation();

  const { complete, failed, total } = data;

  const percentages = {
    complete: (complete / total) * 100,
    failed: (failed / total) * 100,
  };

  const chartData = [
    {
      name: "failed",
      count: failed,
    },
    {
      name: "complete",
      count: complete,
    },
  ];

  const valueFormatter = (number: number) => number.toString();

  return (
    <div className="flex flex-col items-center py-4">
      <DonutChart
        className="h-36 w-36 pb-2"
        showAnimation={false}
        showLabel={false}
        data={chartData}
        category="count"
        index="name"
        valueFormatter={valueFormatter}
        colors={["red", "emerald"]}
      />
      <div className="flex justify-evenly gap-2 lg:gap-3">
        <div>
          {t("pages.explorer.tree.workflow.overview.metrics.successful", {
            percentage: percentages.complete.toFixed(0),
          })}
        </div>
        <div>
          {t("pages.explorer.tree.workflow.overview.metrics.failed", {
            percentage: percentages.failed.toFixed(0),
          })}
        </div>
      </div>
    </div>
  );
};

export default Donut;
