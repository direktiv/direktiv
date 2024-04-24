import { MetricsObjectSchemaType } from "~/api/metrics/schema";

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

export const extractMetrics = (
  data: MetricsObjectSchemaType
): MetricsReport => {
  const { complete, failed, crashed } = data;

  const count = complete + failed + crashed;

  const aggregateFailed = failed + crashed;

  if (count === 0) {
    return { count };
  }

  const items: MetricsItem[] = [
    {
      name: "failed",
      count: aggregateFailed,
      percentage: (aggregateFailed / count) * 100,
    },
    {
      name: "complete",
      count: complete,
      percentage: (complete / count) * 100,
    },
  ];

  return {
    count,
    items,
  };
};
