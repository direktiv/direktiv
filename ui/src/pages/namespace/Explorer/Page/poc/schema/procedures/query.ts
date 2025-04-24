import { KeyValue } from "../primitives/keyValue";
import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const Query = z.object({
  id: z.string().min(1),
  endpoint: TemplateString,
  queryParams: z.array(KeyValue).optional(),
});

export type QueryType = z.infer<typeof Query>;
