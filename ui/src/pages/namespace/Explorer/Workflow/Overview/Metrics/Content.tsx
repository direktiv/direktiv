import Donut from "./Donut";
import { MetricsResponseSchemaType } from "~/api/metrics/schema";
import { NoResult } from "~/design/Table";
import { PieChart } from "lucide-react";
import { getDonutConfig } from "./utils";
import { useTranslation } from "react-i18next";

const Content = ({
  isFetched,
  data,
}: {
  isFetched: boolean;
  data: MetricsResponseSchemaType;
}) => {
  const { t } = useTranslation();
  if (!isFetched)
    return (
      <NoResult icon={PieChart}>
        {t("pages.explorer.tree.workflow.overview.metrics.loading")}
      </NoResult>
    );

  const config = getDonutConfig(data.data);

  if (config.total === 0)
    return (
      <NoResult icon={PieChart}>
        {t("pages.explorer.tree.workflow.overview.metrics.noResult")}
      </NoResult>
    );

  return <Donut config={config} />;
};

export default Content;
