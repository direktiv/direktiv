import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import ScrollContainer from "./ScrollContainer";
import { ScrollText } from "lucide-react";
import { generateLogEntryForClipboard } from "./utils";
import { useLogs } from "~/api/logs/query/logs";
import { useTranslation } from "react-i18next";

const Logs = ({ activityId }: { activityId: string }) => {
  const { t } = useTranslation();

  const { data: allLogs = [] } = useLogs({
    activity: activityId,
  });

  const numberOfLogs = allLogs.length;

  const copyValue = allLogs.map(generateLogEntryForClipboard).join("\n") ?? "";

  return (
    <div className="grid grid-rows-[calc(100vh-10rem)]">
      <Card className="relative m-5 grid p-5">
        <div className="mb-5 flex flex-col gap-5 sm:flex-row">
          <h3 className="flex grow items-center gap-x-2 font-medium">
            <ScrollText className="h-5" />
            {t("pages.mirror.activities.detail.logs.title")}
          </h3>
          <ButtonBar>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <div className="flex grow">
                    <CopyButton
                      value={copyValue}
                      buttonProps={{
                        variant: "outline",
                        size: "sm",
                        className: "grow",
                      }}
                    />
                  </div>
                </TooltipTrigger>
                <TooltipContent>
                  {t("pages.monitoring.logs.tooltips.copy")}
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </ButtonBar>
        </div>
        <ScrollContainer activityId={activityId} />
        <div className="flex items-center justify-center pt-2 text-sm text-gray-11 dark:text-gray-dark-11">
          <span className="relative mr-2 flex h-3 w-3">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gray-11 opacity-75 dark:bg-gray-dark-11"></span>
            <span className="relative inline-flex h-3 w-3 rounded-full bg-gray-11 dark:bg-gray-dark-11"></span>
          </span>
          {t("pages.monitoring.logs.logsCount", { count: numberOfLogs })}
        </div>
      </Card>
    </div>
  );
};

export default Logs;
