import { DonutChart, DonutChartProps } from "@tremor/react";

export type DonutConfigType = {
  items: { label: string; count: number }[];
  colors: DonutChartProps["colors"];
  legend: JSX.Element;
};

const Donut = ({ config }: { config: DonutConfigType }) => {
  const valueFormatter = (number: number) => number.toString();

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
      <div className="flex justify-evenly gap-2 lg:gap-3">{config.legend}</div>
    </div>
  );
};

export default Donut;
