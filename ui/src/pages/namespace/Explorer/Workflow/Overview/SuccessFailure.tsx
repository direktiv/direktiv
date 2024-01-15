import { DonutChart } from "@tremor/react";
import { useTranslation } from "react-i18next";

const SuccessFailure = ({
  data,
}: {
  data: { successful: number; failed: number };
}) => {
  const { t } = useTranslation();

  const { successful, failed } = data;
  const total = data.successful + data.failed;

  const percentages = {
    successful: (successful / total) * 100,
    failed: (failed / total) * 100,
  };

  const chartData = [
    {
      name: "failed",
      count: failed,
    },
    {
      name: "successful",
      count: successful,
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
      <div className="flex justify-evenly">
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
    </div>
  );
};

export default SuccessFailure;
