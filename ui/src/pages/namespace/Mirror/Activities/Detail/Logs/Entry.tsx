import {
  ComponentPropsWithoutRef,
  FC,
  PropsWithChildren,
  forwardRef,
} from "react";
import {
  formatLogTime,
  logLevelToLogEntryVariant,
  twMergeClsx,
} from "~/util/helpers";

import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import { useTranslation } from "react-i18next";

type LogSegmentProps = PropsWithChildren & {
  className?: string;
  display: boolean;
};

const LogSegment: FC<LogSegmentProps> = ({ display, className, children }) => {
  if (!display) return <></>;
  return <span className={twMergeClsx("pr-3", className)}>{children}</span>;
};

type LogEntryProps = ComponentPropsWithoutRef<typeof LogEntry>;
type Props = { logEntry: LogEntryType } & LogEntryProps;
export const Entry = forwardRef<HTMLDivElement, Props>(
  ({ logEntry, ...props }, ref) => {
    const { t } = useTranslation();
    const { msg, level, time } = logEntry;
    const timeFormated = formatLogTime(time);

    return (
      <LogEntry
        variant={logLevelToLogEntryVariant(level)}
        time={timeFormated}
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
