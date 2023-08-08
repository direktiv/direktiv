import { ComponentProps, ComponentPropsWithoutRef, forwardRef } from "react";

import { LogEntry } from "~/design/Logs";
import { NamespaceLogSchemaType } from "~/api/namespaces/schema";
import { formatLogTime } from "~/util/helpers";
import { useLogsPreferencesVerboseLogs } from "~/util/store/logs";

type LogEntryVariant = ComponentProps<typeof LogEntry>["variant"];
type logLevel = NamespaceLogSchemaType["level"];

// mage log level a more generic type, also this method
const logLevelToLogEntryVariant = (level: logLevel): LogEntryVariant => {
  switch (level) {
    case "error":
      return "error";
    case "info":
      return "info";
    case "debug":
      return undefined;
    default:
      break;
  }
};

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;

type Props = { logEntry: NamespaceLogSchemaType } & LogEntryProps;

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
        {/* {verbose && tags["loop-index"] && (
          <>
            <span className="opacity-75">{tags["loop-index"]}</span>{" "}
          </>
        )}
        {verbose && tags["workflow"] && (
          <span className="opacity-75">{tags["workflow"]}</span>
        )}
        {verbose && tags["state-id"] && (
          <span className="opacity-60">/{tags["state-id"]}</span>
        )}{" "} */}
        {msg}
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
