import { WorkflowValidationSchema, WorkflowValidationSchemaType } from ".";

import { MarkerSeverity } from "monaco-editor";
import type { editor } from "monaco-editor";
import z from "zod";

type SeverityMapKey = Exclude<
  WorkflowValidationSchemaType[number]["severity"],
  ""
>;

const severityMap: Record<SeverityMapKey, MarkerSeverity> = {
  error: MarkerSeverity.Error,
  warning: MarkerSeverity.Warning,
  info: MarkerSeverity.Info,
  hint: MarkerSeverity.Hint,
};

export const MonacoMarkerSchema = WorkflowValidationSchema.transform(
  (messages): editor.IMarkerData[] =>
    messages.map((item) => {
      const severity =
        item.severity === ""
          ? MarkerSeverity.Error
          : severityMap[item.severity];
      return {
        startLineNumber: item.startLine,
        startColumn: item.startColumn,
        endLineNumber: item.endLine,
        endColumn: item.endColumn,
        message: item.message,
        severity,
      };
    })
);

export type MonacoMarkerSchemaType = z.infer<typeof MonacoMarkerSchema>;
