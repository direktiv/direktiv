import { TemplateString } from "./templateString";
import { z } from "zod";

/**
 * A key-value pair where the key is a simple string and the value is a
 * template string that can contain dynamic placeholders. This is commonly
 * used for headers, query parameters, and other simple data structures
 * that need to support strings containing dynamic values.
 */
export const KeyValue = z.object({
  key: z.string().min(1),
  value: TemplateString,
});

export type KeyValueType = z.infer<typeof KeyValue>;
