import { ComponentPropsWithoutRef, forwardRef } from "react";
import {
  formatLogTime,
  logLevelToLogEntryVariant_DEPRECATED,
} from "~/util/helpers";

import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { useLogsPreferencesVerboseLogs } from "~/util/store/logs";

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;
type Props = { logEntry: LogEntryType } & LogEntryProps;

export const Entry = forwardRef<HTMLDivElement, Props>(
  ({ logEntry, ...props }, ref) => {
    const { msg, level, time, workflow } = logEntry;
    const timeFormated = formatLogTime(time);
    const verbose = useLogsPreferencesVerboseLogs();

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant_DEPRECATED(level)}
        time={timeFormated}
        ref={ref}
        {...props}
      >
        {verbose && workflow && <span className="opacity-75">{workflow}</span>}
        {verbose && workflow && " "}
        {msg}
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
