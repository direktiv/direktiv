import { JsOutboundFormSchema } from "./jsOutbound";
import { z } from "zod";

export const OutboundPluginFormSchema = z.discriminatedUnion("type", [
  JsOutboundFormSchema,
]);

export type OutboundPluginFormSchemaType = z.infer<
  typeof OutboundPluginFormSchema
>;
