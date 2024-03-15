import { LogEntryType } from "~/api/logs/schema";
import { formatLogTime } from "~/util/helpers";

export const generateLogEntryForClipboard = (logEntry: LogEntryType) =>
  `${logEntry.id} - ${formatLogTime(logEntry.time)} - ${logEntry?.level} - ${
    logEntry.msg
  }`;
