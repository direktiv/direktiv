import {
  ComponentPropsWithoutRef,
  FC,
  PropsWithChildren,
  forwardRef,
} from "react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import {
  formatLogTime,
  logLevelToLogEntryVariant,
  twMergeClsx,
} from "~/util/helpers";

import { Link } from "react-router-dom";
import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { pages } from "~/util/router/pages";
import { useInstanceId } from "../../store/instanceContext";
import { useLogsPreferencesVerboseLogs } from "~/util/store/logs";
import { useTranslation } from "react-i18next";

type LogSegmentProps = PropsWithChildren & {
  className?: string;
  display: boolean;
};

const LogSegment: FC<LogSegmentProps> = ({ display, className, children }) => {
  if (!display) return <></>;
  return <span className={twMergeClsx("pr-3", className)}>{children}</span>;
};

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;
type Props = { logEntry: LogEntryType } & LogEntryProps;
export const Entry = forwardRef<HTMLDivElement, Props>(
  ({ logEntry, ...props }, ref) => {
    const instanceId = useInstanceId();
    const { t } = useTranslation();
    const { msg, level, time, workflow, namespace } = logEntry;
    const timeFormated = formatLogTime(time);
    const verbose = useLogsPreferencesVerboseLogs();

    const workflowPath = workflow?.workflow;

    const workflowLink = pages.explorer.createHref({
      path: workflow?.workflow,
      namespace,
      subpage: "workflow",
    });

    const instanceLink = pages.instances.createHref({
      namespace,
      instance: workflow?.instance,
    });

    if (!workflow) return <></>;

    const isChildInstanceEntry = instanceId !== workflow.instance;

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant(level)}
        time={timeFormated}
        ref={ref}
        {...props}
      >
        <LogSegment display={verbose} className="opacity-60">
          {t("pages.instances.detail.logs.entry.stateLabel")} {workflow.state}
        </LogSegment>
        <LogSegment display={true}>
          {t("pages.instances.detail.logs.entry.messageLabel")} {msg}
        </LogSegment>
        <LogSegment display={isChildInstanceEntry}>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Link
                  to={workflowLink}
                  className="underline opacity-60"
                  target="_blank"
                >
                  {workflowPath}
                </Link>
              </TooltipTrigger>
              <TooltipContent side="right">
                {t("pages.instances.detail.logs.entry.workflowTooltip")}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>{" "}
          (
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <Link
                  to={instanceLink}
                  className="underline opacity-60"
                  target="_blank"
                >
                  {t("pages.instances.detail.logs.entry.instanceLabel")}{" "}
                  {workflow.instance.slice(0, 8)}
                </Link>
              </TooltipTrigger>
              <TooltipContent side="right">
                {t("pages.instances.detail.logs.entry.instanceTooltip")}
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
          )
        </LogSegment>
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
