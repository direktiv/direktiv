import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import ApiError from "~/components/ApiError";
import { ButtonBar } from "~/design/ButtonBar";
import CopyButton from "~/design/CopyButton";
import { NoPermissions } from "~/design/Table";
import ScrollContainer from "./Scrollcontainer";
import { ScrollText } from "lucide-react";
import { getMonitoringLogEntryForClipboard } from "~/components/Logs/utils";
import { useLogs } from "~/api/logs/query/logs";
import { useTranslation } from "react-i18next";

const LogsPanel = () => {
  const { t } = useTranslation();

  const {
    data: logLines = [],
    error,
    isFetched,
    isAllowed,
    noPermissionMessage,
  } = useLogs();

  const numberOfLogLines = logLines.length;

  const copyValue =
    logLines.map(getMonitoringLogEntryForClipboard).join("\n") ?? "";

  if (!isFetched) return null;

  if (!isAllowed) return <NoPermissions>{noPermissionMessage}</NoPermissions>;

  return (
    <>
      <div className="mb-3 flex flex-col gap-5 sm:flex-row">
        <h3 className="flex grow gap-x-2 font-medium">
          <ScrollText className="h-5" />
          {t("components.logs.title")}
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
              <TooltipContent>{t("components.logs.copy")}</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </ButtonBar>
      </div>
      <div className="mb-3">
        {error && (
          <ApiError error={error} label={t("pages.monitoring.apiError")} />
        )}
      </div>
      <ScrollContainer />
      <div className="flex items-center justify-center pt-2 text-sm text-gray-11 dark:text-gray-dark-11">
        <span className="relative mr-2 flex h-3 w-3">
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gray-11 opacity-75 dark:bg-gray-dark-11"></span>
          <span className="relative inline-flex h-3 w-3 rounded-full bg-gray-11 dark:bg-gray-dark-11"></span>
        </span>
        {t("components.logs.logsCount", { count: numberOfLogLines })}
      </div>
    </>
  );
};

export default LogsPanel;
