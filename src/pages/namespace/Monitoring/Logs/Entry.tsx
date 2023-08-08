import { ComponentProps, ComponentPropsWithoutRef, forwardRef } from "react";

import { LogEntry } from "~/design/Logs";
import { NamespaceLogSchemaType } from "~/api/namespaces/schema";
import { formatLogTime } from "~/util/helpers";

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
    const { msg, t, level } = logEntry;
    const time = formatLogTime(t);

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant(level)}
        time={time}
        ref={ref}
        {...props}
      >
        {msg}
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
