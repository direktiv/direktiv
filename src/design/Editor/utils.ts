import { editorLanguageSchema } from "./schema";

export const supportedLanguages = [
  "html",
  "css",
  "json",
  "shell",
  "plaintext",
  "yaml",
] as const;

export type EditorLanguagesType = (typeof supportedLanguages)[number];

export const mimeTypeToEditorSyntax = (
  mimeType: string | undefined
): EditorLanguagesType | undefined => {
  const parsed = editorLanguageSchema.safeParse(mimeType);
  if (!parsed.success) return undefined;
  return parsed.data;
};
