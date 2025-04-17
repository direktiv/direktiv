import { ComponentPropsWithoutRef, forwardRef } from "react";
import { formatLogTime, logLevelToLogEntryVariant } from "~/util/helpers";

import { Link } from "@tanstack/react-router";
import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { LogSegment } from "~/components/Logs/LogSegment";
import { useTranslation } from "react-i18next";

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;
type Props = { logEntry: LogEntryType } & LogEntryProps;
export const Entry = forwardRef<HTMLDivElement, Props>(
  ({ logEntry, ...props }, ref) => {
    const { t } = useTranslation();
    const { msg, level, time, workflow, namespace } = logEntry;
    const formattedTime = formatLogTime(time);
    const hasNamespaceInformation = !!namespace;

    const isWorkflowLog = !!workflow;
    const workflowPath = workflow?.workflow;

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant(level)}
        time={formattedTime}
        ref={ref}
        {...props}
      >
        <LogSegment display={true}>
          {t("components.logs.logEntry.messageLabel")} {msg}
        </LogSegment>
        <LogSegment display={isWorkflowLog && hasNamespaceInformation}>
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
            (
            {workflow?.instance && (
              <Link
                to="/n/$namespace/instances/$id"
                params={{
                  namespace: namespace ?? "",
                  id: workflow?.instance ?? "",
                }}
                className="underline"
                target="_blank"
              >
                {t("components.logs.logEntry.instanceLabel")}{" "}
                {workflow?.instance.slice(0, 8)}
              </Link>
            )}
            )
          </span>
        </LogSegment>
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
