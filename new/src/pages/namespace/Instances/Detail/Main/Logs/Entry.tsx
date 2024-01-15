import { ComponentPropsWithoutRef, forwardRef } from "react";
import { formatLogTime, logLevelToLogEntryVariant } from "~/util/helpers";

import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { useLogsPreferencesVerboseLogs } from "~/util/store/logs";

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;
type Props = { logEntry: LogEntryType } & LogEntryProps;

export const Entry = forwardRef<HTMLDivElement, Props>(
  ({ logEntry, ...props }, ref) => {
    const { msg, t, level, tags } = logEntry;
    const time = formatLogTime(t);
    const verbose = useLogsPreferencesVerboseLogs();

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant(level)}
        time={time}
        ref={ref}
        {...props}
      >
        {verbose && tags["loop-index"] && (
          <>
            <span className="opacity-75">{tags["loop-index"]}</span>{" "}
          </>
        )}
        {verbose && tags["workflow"] && (
          <span className="opacity-75">{tags["workflow"]}</span>
        )}
        {verbose && tags["state-id"] && (
          <span className="opacity-60">/{tags["state-id"]}</span>
        )}{" "}
        {msg}
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
