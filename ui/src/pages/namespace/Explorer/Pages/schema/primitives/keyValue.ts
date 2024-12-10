import { DynamicString } from "./dynamicString";
import { z } from "zod";

export const KeyValue = z.object({
  key: z.string().min(1),
  value: DynamicString,
});
