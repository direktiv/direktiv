import { Id } from "../primitives/id";
import { KeyValue } from "../primitives/keyValue";
import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const Query = z.object({
  id: Id,
  endpoint: TemplateString,
  queryParams: z.array(KeyValue).optional(),
});

export type QueryType = z.infer<typeof Query>;
