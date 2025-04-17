import { ComponentPropsWithoutRef, forwardRef } from "react";
import { formatLogTime, logLevelToLogEntryVariant } from "~/util/helpers";

import { Link } from "@tanstack/react-router";
import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { LogSegment } from "~/components/Logs/LogSegment";
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

    const hasWorkflowInformation = !!workflow;

    const workflowPath = workflow?.workflow;

    const isChildInstanceEntry = workflow
      ? instanceId !== workflow.instance
      : false;

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant(level)}
        time={formattedTime}
        ref={ref}
        {...props}
      >
        <LogSegment
          display={verbose && hasWorkflowInformation}
          className="opacity-60"
        >
          {t("components.logs.logEntry.stateLabel")} {workflow?.state}
        </LogSegment>
        <LogSegment display={true}>
          {t("components.logs.logEntry.messageLabel")} {msg}
        </LogSegment>
        <LogSegment display={isChildInstanceEntry && hasWorkflowInformation}>
          <span className="opacity-60">
            <Link
              to="/n/$namespace/explorer/workflow/edit/$"
              params={{
                namespace: namespace ?? "",
                _splat: workflowPath ?? "",
              }}
            >
              {workflowPath}
            </Link>{" "}
            {workflow?.instance && (
              <Link
                to="/n/$namespace/instances/$id"
                params={{ namespace: namespace ?? "", id: workflow?.instance }}
                className="underline"
                target="_blank"
              >
                {t("components.logs.logEntry.instanceLabel")}{" "}
                {workflow?.instance.slice(0, 8)}
              </Link>
            )}
          </span>
        </LogSegment>
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
