import { Bug, Loader2, Maximize2, Minimize2, ScrollText } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import { useFilters, useInstanceId } from "../../store/instanceContext";
import {
  useLogsPreferencesActions,
  useLogsPreferencesMaximizedPanel,
  useLogsPreferencesVerboseLogs,
} from "~/util/store/logs";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import CopyButton from "~/design/CopyButton";
import Filters from "./Filters";
import ScrollContainer from "./ScrollContainer";
import { Toggle } from "~/design/Toggle";
import { formatTime } from "./utils";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useLogs } from "~/api/logs/query/get";
import { useTranslation } from "react-i18next";

const LogsPanel = () => {
  const { t } = useTranslation();
  const { setVerboseLogs, setMaximizedPanel } = useLogsPreferencesActions();

  const instanceId = useInstanceId();
  const filters = useFilters();
  const { data: instanceDetailsData } = useInstanceDetails({ instanceId });
  const { data: logData } = useLogs({
    instanceId,
    filters,
  });

  // get user preferences
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const verboseLogs = useLogsPreferencesVerboseLogs();

  const isMaximized = maximizedPanel === "logs";

  const copyValue =
    logData?.results.map((x) => `${formatTime(x.t)} ${x.msg}`).join("\n") ?? "";

  const resultCount = logData?.results.length ?? 0;
  const isPending = instanceDetailsData?.instance.status === "pending";

  return (
    <>
      <div className="mb-5 flex flex-col gap-5 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 font-medium">
          <ScrollText className="h-5" />
          {t("pages.instances.detail.logs.title")}
        </h3>
        <Filters />
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
            {/* <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Toggle
                    size="sm"
                    className="grow"
                    pressed={wordWrap}
                    onClick={() => {
                      setWordWrap(!wordWrap);
                    }}
                  >
                    <WrapText />
                  </Toggle>
                </div>
              </TooltipTrigger>
              <TooltipContent>
                {t("pages.instances.detail.logs.tooltips.wordWrap")}
              </TooltipContent>
            </Tooltip> */}
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
                {t("pages.instances.detail.logs.tooltips.copy")}
              </TooltipContent>
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
      <div className="flex items-center justify-center pt-2 text-sm text-gray-11">
        {isPending && <Loader2 className="h-3 animate-spin" />}
        {t("pages.instances.detail.logs.logsCount", { count: resultCount })}
      </div>
    </>
  );
};

export default LogsPanel;
