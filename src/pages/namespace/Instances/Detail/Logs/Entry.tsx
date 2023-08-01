import { ComponentProps, ComponentPropsWithoutRef, FC } from "react";

import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import moment from "moment";
import { useLogsPreferencesVerboseLogs } from "~/util/store/logs";

type LogEntryVariant = ComponentProps<typeof LogEntry>["variant"];
type logLevel = LogEntryType["level"];

const logLevelToLogEntryVariant = (level: logLevel): LogEntryVariant => {
  switch (level) {
    case "error":
    case "panic":
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

const Entry: FC<{ logEntry: LogEntryType } & LogEntryProps> = ({
  logEntry,
  ...props
}) => {
  const { msg, t, level, tags } = logEntry;
  const time = moment(t).format("HH:mm:ss");
  const verbose = useLogsPreferencesVerboseLogs();

  return (
    <LogEntry variant={logLevelToLogEntryVariant(level)} time={time} {...props}>
      {verbose && tags["loop-index"] && (
        <>
          <span className="opacity-75">{tags["loop-index"]}</span>{" "}
        </>
      )}
      {verbose && tags["workflow"] && (
        <span className="opacity-75">{tags["workflow"]}</span>
      )}
      {verbose && tags["workflow"] && (
        <span className="opacity-60">/{tags["state-id"]}</span>
      )}{" "}
      {msg}
    </LogEntry>
  );
};

export default Entry;
