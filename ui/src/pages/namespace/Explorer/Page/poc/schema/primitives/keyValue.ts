import { TemplateString } from "./templateString";
import { z } from "zod";

export const KeyValue = z.object({
  key: z.string().min(1),
  value: TemplateString,
});

export type KeyValueType = z.infer<typeof KeyValue>;
