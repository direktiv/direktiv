import { ComponentPropsWithoutRef, forwardRef } from "react";
import { formatLogTime, logLevelToLogEntryVariant } from "~/util/helpers";

import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { LogSegment } from "~/components/Logs/LogSegment";
import { useTranslation } from "react-i18next";

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;
type Props = { logEntry: LogEntryType } & LogEntryProps;

export const Entry = forwardRef<HTMLDivElement, Props>(
  ({ logEntry, ...props }, ref) => {
    const { t } = useTranslation();
    const { msg, level, time } = logEntry;
    const formattedTime = formatLogTime(time);

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
      </LogEntry>
    );
  }
);

Entry.displayName = "Entry";

export default Entry;
