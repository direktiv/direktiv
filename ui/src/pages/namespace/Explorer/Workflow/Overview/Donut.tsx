import { DonutChart } from "@tremor/react";
import { MetricsReport } from "./utils";
import { NoResult } from "~/design/Table";
import { PieChart } from "lucide-react";
import { useTranslation } from "react-i18next";

const Donut = ({ metrics }: { metrics: MetricsReport }) => {
  const { t } = useTranslation();

  const valueFormatter = (number: number) => number.toString();

  if (!metrics.items)
    return (
      <NoResult icon={PieChart}>
        {t("pages.explorer.tree.workflow.overview.metrics.noResult")}
      </NoResult>
    );

  const { items } = metrics;

  const chartData = items.map((item) => ({
    name: item.name,
    count: item.count,
  }));

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
        {items.map((item) => (
          <div key={item.name}>
            {t(`pages.explorer.tree.workflow.overview.metrics.${item.name}`, {
              percentage: item.percentage.toFixed(0),
            })}
          </div>
        ))}
      </div>
    </div>
  );
};

export default Donut;
