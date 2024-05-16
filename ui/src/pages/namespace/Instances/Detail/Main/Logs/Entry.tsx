import { ComponentPropsWithoutRef, forwardRef } from "react";
import { formatLogTime, logLevelToLogEntryVariant } from "~/util/helpers";

import { Link } from "react-router-dom";
import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { LogSegment } from "~/components/Logs/LogSegment";
import { useInstanceId } from "../../store/instanceContext";
import { useLogsPreferencesVerboseLogs } from "~/util/store/logs";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;
type Props = { logEntry: LogEntryType } & LogEntryProps;

export const Entry = forwardRef<HTMLDivElement, Props>(
  ({ logEntry, ...props }, ref) => {
    const pages = usePages();
    const instanceId = useInstanceId();
    const { t } = useTranslation();
    const { msg, error, level, time, workflow, namespace } = logEntry;
    const formattedTime = formatLogTime(time);
    const verbose = useLogsPreferencesVerboseLogs();

    const workflowPath = workflow?.workflow;

    const hasWorkflowInformation = !!workflow;

    const workflowLink = pages.explorer.createHref({
      path: workflow?.workflow,
      namespace: namespace ?? "",
      subpage: "workflow",
    });

    const instanceLink = pages.instances.createHref({
      namespace: namespace ?? "",
      instance: workflow?.instance,
    });

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
        <LogSegment display={error ? true : false}>
          {t("components.logs.logEntry.errorLabel")} {error}
        </LogSegment>
        <LogSegment display={isChildInstanceEntry && hasWorkflowInformation}>
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
