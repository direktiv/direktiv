import { ExtendedKeyValue } from "../primitives/extendedKeyValue";
import { KeyValue } from "../primitives/keyValue";
import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const mutationMethods = ["POST", "PUT", "PATCH", "DELETE"] as const;

const MutationMethod = z.enum(mutationMethods);

export const Mutation = z.object({
  method: MutationMethod,
  url: TemplateString.min(1),
  queryParams: z.array(KeyValue).optional(),
  requestHeaders: z.array(KeyValue).optional(),
  requestBody: z.array(ExtendedKeyValue).optional(),
});

export type MutationType = z.infer<typeof Mutation>;
