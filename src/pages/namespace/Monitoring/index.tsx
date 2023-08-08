import { ActivitySquare } from "lucide-react";
import { Card } from "~/design/Card";
import Logs from "./Logs";
import { twMergeClsx } from "~/util/helpers";
import { useNamespaceLogsStream } from "~/api/namespaces/query/logs";
import { useTranslation } from "react-i18next";

const InstancesPage = () => {
  const { t } = useTranslation();
  /**
   * the streaming hooks will update the react query cache
   * when it received new data. This will trigger a rerender
   * of all components that consume this data via useQuery.
   *
   * This is why it's important to place this hook in a separate
   * parent component on top of the consuming components. It
   * will ensure that the hook will not rerun itself (canceling
   * the stream and restarting a new one) when it updates the cache
   */
  useNamespaceLogsStream();

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <ActivitySquare className="h-5" />
        {t("pages.monitoring.title")}
      </h3>
      <div
        className={twMergeClsx(
          "grid grow gap-5",
          "grid-rows-[100vh_50vh_50vh]",
          "md:grid-rows-[minmax(300px,45vh)_1fr]",
          "md:grid-cols-[minmax(430px,1fr)_1fr]"
        )}
      >
        <Card className="relative grid grid-rows-[auto,1fr,auto] p-5 md:col-span-2">
          <Logs />
        </Card>
        <Card className="p-5"></Card>
        <Card className="p-5"></Card>
      </div>
    </div>
  );
};

export default InstancesPage;
