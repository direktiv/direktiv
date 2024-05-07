import { Bug, Maximize2, Minimize2, ScrollText } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import {
  useLogsPreferencesActions,
  useLogsPreferencesMaximizedPanel,
  useLogsPreferencesVerboseLogs,
} from "~/util/store/logs";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import CopyButton from "~/design/CopyButton";
import ScrollContainer from "./ScrollContainer";
import { Toggle } from "~/design/Toggle";
import { getInstanceLogEntryForClipboard } from "~/components/Logs/utils";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "../../store/instanceContext";
import { useLogs } from "~/api/logs/query/logs";
import { useTranslation } from "react-i18next";

const LogsPanel = () => {
  const { t } = useTranslation();
  const { setVerboseLogs, setMaximizedPanel } = useLogsPreferencesActions();

  const instanceId = useInstanceId();

  const { data: instanceDetailsData } = useInstanceDetails({ instanceId });

  const { data: logLines = [] } = useLogs({
    instance: instanceId,
  });

  // get user preferences
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const verboseLogs = useLogsPreferencesVerboseLogs();

  const isMaximized = maximizedPanel === "logs";

  const isPending = instanceDetailsData?.status === "pending";

  const numberOfLogLines = logLines.length;

  const copyValue =
    logLines.map(getInstanceLogEntryForClipboard).join("\n") ?? "";

  return (
    <>
      <div className="mb-5 flex flex-col gap-5 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 font-medium">
          <ScrollText className="h-5" />
          {t("pages.instances.detail.logs.title", {
            path: instanceDetailsData?.path,
          })}
        </h3>
        <ButtonBar>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Toggle
                    size="sm"
                    className="grow"
                    pressed={verboseLogs}
                    onClick={() => {
                      setVerboseLogs(!verboseLogs);
                    }}
                  >
                    <Bug />
                  </Toggle>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                {t("pages.instances.detail.logs.tooltips.verbose")}
              </TooltipContent>
            </Tooltip>
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
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Button
                    icon
                    variant="outline"
                    size="sm"
                    className="grow"
                    onClick={() => {
                      setMaximizedPanel(isMaximized ? "none" : "logs");
                    }}
                  >
                    {isMaximized ? <Minimize2 /> : <Maximize2 />}
                  </Button>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                {isMaximized
                  ? t("pages.instances.detail.logs.tooltips.minimize")
                  : t("pages.instances.detail.logs.tooltips.maximize")}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </ButtonBar>
      </div>
      <ScrollContainer />
      <div className="flex items-center justify-center pt-2 text-sm text-gray-11 dark:text-gray-dark-11">
        {isPending && (
          <span className="relative mr-2 flex h-3 w-3">
            <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-gray-11 opacity-75 dark:bg-gray-dark-11"></span>
            <span className="relative inline-flex h-3 w-3 rounded-full bg-gray-11 dark:bg-gray-dark-11"></span>
          </span>
        )}
        {t("components.logs.logsCount", { count: numberOfLogLines })}
      </div>
    </>
  );
};

export default LogsPanel;
