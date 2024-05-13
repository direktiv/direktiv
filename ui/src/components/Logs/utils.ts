import { LogEntryType } from "~/api/logs/schema";
import { formatLogTime } from "~/util/helpers";

export const getInstanceLogEntryForClipboard = (logEntry: LogEntryType) =>
  createLogEntryForClipboard([
    logEntry.id,
    formatLogTime(logEntry.time),
    logEntry?.level,
    logEntry?.workflow?.workflow,
    logEntry?.workflow?.state,
    logEntry.msg,
    logEntry.error ?? undefined,
  ]);

export const getMirrorLogEntryForClipboard = (logEntry: LogEntryType) =>
  createLogEntryForClipboard([
    logEntry.id,
    formatLogTime(logEntry.time),
    logEntry?.level,
    logEntry?.msg,
    logEntry.error ?? undefined,
  ]);

export const getMonitoringLogEntryForClipboard = (logEntry: LogEntryType) => {
  const isWorkflowLog = !!logEntry.workflow;
  const workflowInfos = isWorkflowLog
    ? [logEntry.workflow?.workflow, logEntry.workflow?.instance]
    : [];

  return createLogEntryForClipboard([
    logEntry.id,
    formatLogTime(logEntry.time),
    logEntry?.level,
    logEntry.msg,
    logEntry.error ?? undefined,
    ...workflowInfos,
  ]);
};

export const getRouteLogEntryForClipboard = (logEntry: LogEntryType) =>
  createLogEntryForClipboard([
    logEntry.id,
    formatLogTime(logEntry.time),
    logEntry?.level,
    logEntry?.route?.path,
    logEntry.msg,
    logEntry.error ?? undefined,
  ]);

const createLogEntryForClipboard = (parts: (string | number | undefined)[]) =>
  parts.filter((entry) => entry !== undefined).join(" - ");
