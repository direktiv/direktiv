import { Id } from "../primitives/id";
import { KeyValue } from "../primitives/keyValue";
import { RequestBodySchema } from "./requestBody";
import { TemplateString } from "../primitives/templateString";
import { z } from "zod";

export const mutationMethods = ["POST", "PUT", "PATCH", "DELETE"] as const;

const MutationMethod = z.enum(mutationMethods);

export const Mutation = z.object({
  id: Id,
  method: MutationMethod,
  url: TemplateString,
  queryParams: z.array(KeyValue).optional(),
  requestHeaders: z.array(KeyValue).optional(),
  requestBody: z.array(RequestBodySchema).optional(),
});

export type MutationType = z.infer<typeof Mutation>;
