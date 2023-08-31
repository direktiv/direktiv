import { supportedLanguages } from "./utils";
import { z } from "zod";

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
