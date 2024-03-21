import { ActivitySquare } from "lucide-react";
import { Card } from "~/design/Card";
import { Instances } from "./Instances";
import { LogStreamingSubscriber } from "~/api/logs/query/LogStreamingSubscriber";
import Logs from "./Logs";
import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

const InstancesPage = () => {
  const { t } = useTranslation();
  return (
    <>
      <LogStreamingSubscriber />
      <div className="flex grow flex-col gap-y-4 p-5">
        <h3 className="flex items-center gap-x-2 font-bold">
          <ActivitySquare className="h-5" />
          {t("pages.monitoring.title")}
        </h3>
        <div
          className={twMergeClsx(
            "grid grow gap-5",
            "grid-rows-[100vh_50vh_50vh]",
            "md:grid-rows-[55vh_24vh]",
            "md:grid-cols-[1fr_1fr]"
          )}
        >
          <Card className="relative grid grid-rows-[auto,1fr,auto] p-5 md:col-span-2">
            <Logs />
          </Card>
          <Instances />
        </div>
      </div>
    </>
  );
};

export default InstancesPage;
