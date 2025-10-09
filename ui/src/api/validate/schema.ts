import type { MarkerSeverity, editor } from "monaco-editor";

import z from "zod";

export const WorkflowValidationSchema = z.array(
  z.object({
    message: z.string().min(1),
    startLine: z.number(),
    startColumn: z.number(),
    endLine: z.number(),
    endColumn: z.number(),
    severity: z.enum(["hint", "info", "warning", "error", ""]),
  })
);

// Runtime access to constants MarkerSeverity.Error etc. breaks unit
// tests, because monaco editor only exists in a browser environment.
const severityMap: Record<string, MarkerSeverity> = {
  error: 8,
  warning: 4,
  info: 2,
  hint: 1,
};

export const MonacoMarkerSchema = WorkflowValidationSchema.transform(
  (messages): editor.IMarkerData[] =>
    messages.map((item) => ({
      startLineNumber: item.startLine,
      startColumn: item.startColumn,
      endLineNumber: item.endLine,
      endColumn: item.endColumn,
      message: item.message,
      severity: severityMap[item.severity] || 8,
    }))
);
