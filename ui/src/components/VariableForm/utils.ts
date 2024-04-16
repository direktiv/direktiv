import { z } from "zod";

export const mimeTypes = [
  { label: "JSON", value: "application/json" },
  { label: "YAML", value: "application/yaml" },
  { label: "shell", value: "application/x-sh" },
  { label: "plaintext", value: "text/plain" },
  { label: "HTML", value: "text/html" },
  { label: "CSS", value: "text/css" },
];

export const mimeTypeToLanguageDict = {
  "application/json": "json",
  "application/yaml": "yaml",
  "application/x-sh": "shell",
  "text/plain": "plaintext",
  "text/html": "html",
  "text/css": "css",
} as const;

export const getLanguageFromMimeType = (mimeType: string) => {
  const parsed = EditorMimeTypeSchema.safeParse(mimeType);
  if (parsed.success) {
    return mimeTypeToLanguageDict[parsed.data];
  }
  return undefined;
};

export const EditorMimeTypeSchema = z.enum([
  "application/json",
  "application/yaml",
  "application/x-sh",
  "text/plain",
  "text/html",
  "text/css",
]);

export type TextMimeTypeType = z.infer<typeof EditorMimeTypeSchema>;

export const isMimeTypeEditable = (
  mimeType: string
): mimeType is TextMimeTypeType => {
  const parsedMimetype = EditorMimeTypeSchema.safeParse(mimeType);
  return parsedMimetype.success;
};
