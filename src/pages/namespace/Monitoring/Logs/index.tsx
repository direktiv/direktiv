import { Loader2, ScrollText } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";

import { ButtonBar } from "~/design/ButtonBar";
import CopyButton from "~/design/CopyButton";
import ScrollContainer from "./Scrollcontainer";
import { formatLogTime } from "~/util/helpers";
import { useNamespacelogs } from "~/api/namespaces/query/logs";
import { useTranslation } from "react-i18next";

const LogsPanel = () => {
  const { t } = useTranslation();
  const { data } = useNamespacelogs();

  const copyValue =
    data?.results.map((x) => `${formatLogTime(x.t)} ${x.msg}`).join("\n") ?? "";

  const resultCount = data?.results.length ?? 0;

  return (
    <>
      <div className="mb-5 flex flex-col gap-5 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 font-medium">
          <ScrollText className="h-5" />
          {t("pages.monitoring.logs.title")}
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
      <ScrollContainer />
      <div className="flex items-center justify-center pt-2 text-sm text-gray-11">
        <Loader2 className="h-3 animate-spin" />
        {t("pages.monitoring.logs.logsCount", { count: resultCount })}
      </div>
    </>
  );
};

export default LogsPanel;
