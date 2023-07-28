import { ComponentProps, ComponentPropsWithoutRef, FC } from "react";

import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import moment from "moment";

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
  const { msg, t, level } = logEntry;

  const time = moment(t).format("HH:mm:ss");

  return (
    <LogEntry variant={logLevelToLogEntryVariant(level)} time={time} {...props}>
      {msg}
    </LogEntry>
  );
};

export default Entry;
