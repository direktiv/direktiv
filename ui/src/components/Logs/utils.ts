import { LogEntryType } from "~/api/logs/schema";
import { formatLogTime } from "~/util/helpers";

export const getInstanceLogEntryForClipboard = (logEntry: LogEntryType) =>
  createLogEntryForClipboard([
    formatLogTime(logEntry.time),
    logEntry?.level,
    logEntry?.workflow?.workflow,
    logEntry?.workflow?.state ?? undefined,
    logEntry.msg,
  ]);

export const getMirrorLogEntryForClipboard = (logEntry: LogEntryType) =>
  createLogEntryForClipboard([
    formatLogTime(logEntry.time),
    logEntry?.level,
    logEntry?.msg,
  ]);

export const getMonitoringLogEntryForClipboard = (logEntry: LogEntryType) => {
  const isWorkflowLog = !!logEntry.workflow;
  const workflowInfos = isWorkflowLog
    ? [logEntry.workflow?.workflow, logEntry.workflow?.instance]
    : [];

  return createLogEntryForClipboard([
    formatLogTime(logEntry.time),
    logEntry?.level,
    logEntry.msg,
    ...workflowInfos,
  ]);
};

export const getRouteLogEntryForClipboard = (logEntry: LogEntryType) =>
  createLogEntryForClipboard([
    formatLogTime(logEntry.time),
    logEntry?.level,
    logEntry?.route?.path,
    logEntry.msg,
  ]);

const createLogEntryForClipboard = (parts: (string | number | undefined)[]) =>
  parts.filter((entry) => entry !== undefined).join(" - ");
