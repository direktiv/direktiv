import { Bug, Maximize2, Minimize2, WrapText } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import { useFilters, useInstanceId } from "../state/instanceContext";
import {
  useLogsPreferencesActions,
  useLogsPreferencesMaximizedPanel,
  useLogsPreferencesVerboseLogs,
  useLogsPreferencesWordWrap,
} from "~/util/store/logs";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import CopyButton from "~/design/CopyButton";
import Filters from "./Filters";
import ScrollContainer from "./ScrollContainer";
import { Toggle } from "~/design/Toggle";
import { useLogs } from "~/api/logs/query/get";
import { useTranslation } from "react-i18next";

const LogsPanel = () => {
  const { t } = useTranslation();
  const wordWrap = useLogsPreferencesWordWrap();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const verboseLogs = useLogsPreferencesVerboseLogs();
  const { setVerboseLogs, setWordWrap, setMaximizedPanel } =
    useLogsPreferencesActions();

  const instanceId = useInstanceId();
  const filters = useFilters();
  const { data } = useLogs({
    instanceId,
    filters,
  });

  const isMaximized = maximizedPanel === "logs";

  const copyValue = data?.results.map((x) => `${x.msg}`).join("\n") ?? "";

  return (
    <>
      <div className="mb-5 flex gap-x-5">
        <h3 className="grow font-medium">
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
                  <Toggle
                    size="sm"
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
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <CopyButton
                    value={copyValue}
                    buttonProps={{
                      variant: "outline",
                      size: "sm",
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
    </>
  );
};

export default LogsPanel;
