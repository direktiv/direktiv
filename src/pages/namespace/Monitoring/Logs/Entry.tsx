import { ComponentPropsWithoutRef, forwardRef } from "react";
import { formatLogTime, logLevelToLogEntryVariant } from "~/util/helpers";

import { LogEntry } from "~/design/Logs";
import { NamespaceLogSchemaType } from "~/api/namespaces/schema";

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
