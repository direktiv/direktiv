import { ComponentPropsWithoutRef, forwardRef } from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import { formatLogTime, logLevelToLogEntryVariant } from "~/util/helpers";

import { Link } from "react-router-dom";
import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { pages } from "~/util/router/pages";
import { useLogsPreferencesVerboseLogs } from "~/util/store/logs";
import { useTranslation } from "react-i18next";

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;
type Props = { logEntry: LogEntryType } & LogEntryProps;

export const Entry = forwardRef<HTMLDivElement, Props>(
  ({ logEntry, ...props }, ref) => {
    const { t } = useTranslation();
    const { msg, level, time, workflow, namespace } = logEntry;
    const timeFormated = formatLogTime(time);
    const verbose = useLogsPreferencesVerboseLogs();

    const link = pages.explorer.createHref({
      path: workflow?.workflow,
      namespace,
      subpage: "workflow",
    });

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant(level)}
        time={timeFormated}
        ref={ref}
        {...props}
      >
        {verbose && workflow && (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Link
                  to={link}
                  className=" underline opacity-75"
                  target="_blank"
                >
                  {workflow?.workflow}
                </Link>
              </TooltipTrigger>
              <TooltipContent>
                {t("pages.instances.detail.logs.entry.workflowTooltip")}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        )}
        {verbose && workflow && " "}
        {msg}
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
