import { EditorLanguagesType } from ".";
import { z } from "zod";

export const supportedLanguages = [
  "html",
  "css",
  "json",
  "shell",
  "plaintext",
  "yaml",
] as const;

export const mimeTypeToEditorSyntax = (
  mimeType: string | undefined
): EditorLanguagesType | undefined => {
  const parsed = editorLanguageSchema.safeParse(mimeType);
  if (!parsed.success) return undefined;
  return parsed.data;
};

export const editorLanguageSchema = z
  .string()
  .transform((val) => {
    if (val.startsWith("text/html")) return "html";
    if (val.startsWith("text/css")) return "css";
    if (val.startsWith("application/json")) return "json";
    if (val.startsWith("application/x-csh")) return "sh";
    if (val.startsWith("application/x-sh")) return "sh";
    if (val.startsWith("text/")) return "plaintext";
    return val;
  })
  .pipe(z.enum(supportedLanguages));
