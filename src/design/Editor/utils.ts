import { EditorLanguagesType } from ".";
import { z } from "zod";

export const supportedLanguages = [
  "html",
  "css",
  "json",
  "shell",
  "shell",
  "javascript",
  "yaml",
] as const;

export const mimeTypeToEditorSyntax = (
  mimeType: string | undefined
): EditorLanguagesType | undefined => {
  const parsed = editorLanguageSchema.safeParse(mimeType);
  if (!parsed.success) return undefined;
  return parsed.data;
};

/**
 * reference for common mime types
 * https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
 */
export const editorLanguageSchema = z
  .string()
  .transform((val) => {
    switch (val) {
      case "text/html":
        return "html";
      case "text/css":
        return "css";
      case "application/javascript":
        return "javascript";
      case "application/json":
        return "json";
      case "application/x-sh":
      case "application/x-csh":
        return "shell";
      case "text/yaml":
      case "application/direktiv":
        return "yaml";
      default:
        if (val.startsWith("text/")) {
          return "plaintext";
        }
        return val;
    }
  })
  .pipe(z.enum(supportedLanguages));
