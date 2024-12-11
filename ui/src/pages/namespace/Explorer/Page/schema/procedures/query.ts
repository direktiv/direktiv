import { DynamicString } from "../primitives/dynamicString";
import { KeyValue } from "../primitives/keyValue";
import { z } from "zod";

export const Query = z.object({
  id: z.string().min(1),
  endpoint: DynamicString,
  queryParams: KeyValue.optional(),
});

export type QueryType = z.infer<typeof Query>;
