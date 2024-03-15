import { LogEntryType } from "~/api/logs/schema";
import { formatLogTime } from "~/util/helpers";

export const generateClipboardLogEntry = (logEntry: LogEntryType) => {
  const isWorkflowLog = !!logEntry.workflow;

  const worfklowLogInfo = isWorkflowLog
    ? ` - ${logEntry.workflow?.workflow} - ${logEntry.workflow?.instance}`
    : "";

  return `${logEntry.id} - ${formatLogTime(logEntry.time)} - ${
    logEntry?.level
  } - ${logEntry.msg}${worfklowLogInfo}`;
};
