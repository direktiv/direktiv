import { DonutConfigType } from "./Donut";
import { MetricsObjectSchemaType } from "~/api/metrics/schema";
import { Trans } from "react-i18next";

// only the names processed in the component are allowed
type MetricsItem = {
  name: "failed" | "complete";
  count: number;
  percentage: number;
};

export type MetricsReport = {
  count: number;
  items?: MetricsItem[];
};

export const getDonutConfig = (
  data: MetricsObjectSchemaType
): DonutConfigType => {
  // items are hard coded to have easier control over their order,
  // which should correspond to the order of colors
  const items: DonutConfigType["items"] = [
    { label: "complete", count: data.complete },
    { label: "failed", count: data.failed },
    { label: "crashed", count: data.crashed },
  ];

  const colors: DonutConfigType["colors"] = ["emerald", "red", "orange"];

  const total = items.reduce((sum, item) => sum + item.count, 0);

  const successRate = ((data.complete / total) * 100).toFixed(0);

  const legend = (
    <Trans
      i18nKey="pages.explorer.tree.workflow.overview.metrics.legend"
      values={{ successRate }}
    />
  );

  return {
    items,
    legend,
    colors,
  };
};
