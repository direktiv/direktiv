import { DynamicString } from "../primitives/dynamicString";
import { MutationMethod } from "./request/methods";
import { RequestBody } from "./request/requestBody";
import { RequestHeaders } from "./request/requestHeaders";
import { SearchParms } from "./request/searchParams";
import { z } from "zod";

export const Mutation = z.object({
  id: z.string().min(1),
  method: MutationMethod,
  endpoint: DynamicString,
  searchParms: SearchParms,
  headers: RequestHeaders,
  body: RequestBody,
});

export type MutationType = z.infer<typeof Mutation>;
