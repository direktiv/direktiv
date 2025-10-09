import { MarkerSeverity } from "monaco-editor";
import { WorkflowValidationSchema } from ".";
import type { editor } from "monaco-editor";

const severityMap: Record<string, MarkerSeverity> = {
  error: MarkerSeverity.Error,
  warning: MarkerSeverity.Warning,
  info: MarkerSeverity.Info,
  hint: MarkerSeverity.Hint,
};

export const MonacoMarkerSchema = WorkflowValidationSchema.transform(
  (messages): editor.IMarkerData[] =>
    messages.map((item) => ({
      startLineNumber: item.startLine,
      startColumn: item.startColumn,
      endLineNumber: item.endLine,
      endColumn: item.endColumn,
      message: item.message,
      severity: severityMap[item.severity] || MarkerSeverity.Error,
    }))
);
