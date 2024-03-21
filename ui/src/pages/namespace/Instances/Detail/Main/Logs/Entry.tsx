import { ComponentPropsWithoutRef, forwardRef } from "react";
import { formatLogTime, logLevelToLogEntryVariant } from "~/util/helpers";

import { Link } from "react-router-dom";
import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { LogSegment } from "~/components/Logs/LogSegment";
import { pages } from "~/util/router/pages";
import { useInstanceId } from "../../store/instanceContext";
import { useLogsPreferencesVerboseLogs } from "~/util/store/logs";
import { useTranslation } from "react-i18next";

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;
type Props = { logEntry: LogEntryType } & LogEntryProps;

export const Entry = forwardRef<HTMLDivElement, Props>(
  ({ logEntry, ...props }, ref) => {
    const instanceId = useInstanceId();
    const { t } = useTranslation();
    const { msg, level, time, workflow, namespace } = logEntry;
    const formattedTime = formatLogTime(time);
    const verbose = useLogsPreferencesVerboseLogs();

    const workflowPath = workflow?.workflow;

    if (!workflow) return <></>;
    if (!namespace) return <></>;

    const workflowLink = pages.explorer.createHref({
      path: workflow?.workflow,
      namespace,
      subpage: "workflow",
    });

    const instanceLink = pages.instances.createHref({
      namespace,
      instance: workflow?.instance,
    });

    const isChildInstanceEntry = instanceId !== workflow.instance;

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant(level)}
        time={formattedTime}
        ref={ref}
        {...props}
      >
        <LogSegment display={verbose} className="opacity-60">
          {t("components.logs.logEntry.stateLabel")} {workflow.state}
        </LogSegment>
        <LogSegment display={true}>
          {t("components.logs.logEntry.messageLabel")} {msg}
        </LogSegment>
        <LogSegment display={isChildInstanceEntry}>
          <span className="opacity-60">
            <Link to={workflowLink} className="underline" target="_blank">
              {workflowPath}
            </Link>{" "}
            (
            <Link to={instanceLink} className="underline" target="_blank">
              {t("components.logs.logEntry.instanceLabel")}{" "}
              {workflow.instance.slice(0, 8)}
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
