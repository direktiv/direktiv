import { DonutChart, DonutChartProps } from "@tremor/react";

import { MetricsReport } from "./utils";
import { NoResult } from "~/design/Table";
import { PieChart } from "lucide-react";
import { useTranslation } from "react-i18next";

export type DonutConfigType = {
  items: { label: string; count: number }[];
  colors: DonutChartProps["colors"];
  legend: JSX.Element;
};

const Donut = ({ config }: { config: DonutConfigType }) => {
  const { t } = useTranslation();

  // new: all data in one object
  // const config: DonutConfigType = {
  //   items: [
  //     { label: "complete", count: 5 },
  //     { label: "failed", count: 2 },
  //     { label: "crashed", count: 3 },
  //     { label: "cancelled", count: 2 },
  //   ],
  //   colors: ["emerald", "red", "orange", "stone"],
  // };

  const valueFormatter = (number: number) => number.toString();

  // if (!metrics.items)
  //   return (
  //     <NoResult icon={PieChart}>
  //       {t("pages.explorer.tree.workflow.overview.metrics.noResult")}
  //     </NoResult>
  //   );

  // const { items } = metrics;

  const chartData = config.items.map((item) => ({
    name: item.label,
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
        colors={config.colors}
      />
      <div className="flex justify-evenly gap-2 lg:gap-3">
        {config.legend}
        {/* {items.map((item) => (
          <div key={item.name}>
            {t(`pages.explorer.tree.workflow.overview.metrics.${item.name}`, {
              percentage: item.percentage.toFixed(0),
            })}
          </div>
        ))} */}
      </div>
    </div>
  );
};

export default Donut;
