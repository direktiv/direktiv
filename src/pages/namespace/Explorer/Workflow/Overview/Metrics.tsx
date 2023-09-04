import { DonutChart } from "@tremor/react";
import { useTranslation } from "react-i18next";

const Metrics = ({
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
    <>
      <DonutChart
        showAnimation={false}
        showLabel={false}
        className="mt-6"
        data={chartData}
        category="count"
        index="name"
        valueFormatter={valueFormatter}
        colors={["red", "emerald"]}
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
    </>
  );
};

export default Metrics;
