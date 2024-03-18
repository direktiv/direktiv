import {
  ComponentPropsWithoutRef,
  FC,
  PropsWithChildren,
  forwardRef,
} from "react";
import {
  formatLogTime,
  logLevelToLogEntryVariant,
  twMergeClsx,
} from "~/util/helpers";

import { Link } from "react-router-dom";
import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { pages } from "~/util/router/pages";
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
    const { t } = useTranslation();
    const { msg, level, time, workflow, namespace } = logEntry;
    const timeFormated = formatLogTime(time);

    if (!namespace) return <></>;

    const isWorkflowLog = !!workflow;
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

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant(level)}
        time={timeFormated}
        ref={ref}
        {...props}
      >
        <LogSegment display={true}>
          {t("components.logs.logEntry.messageLabel")} {msg}
        </LogSegment>
        <LogSegment display={isWorkflowLog}>
          <span className="opacity-60">
            <Link to={workflowLink} className="underline" target="_blank">
              {workflowPath}
            </Link>{" "}
            (
            <Link to={instanceLink} className="underline" target="_blank">
              {t("components.logs.logEntry.instanceLabel")}{" "}
              {workflow?.instance.slice(0, 8)}
            </Link>
            )
          </span>
        </LogSegment>
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
