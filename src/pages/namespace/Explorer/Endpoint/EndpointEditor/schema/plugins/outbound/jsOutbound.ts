import { outboundPluginTypes } from ".";
import { z } from "zod";

export const JsOutboundFormSchema = z.object({
  type: z.literal(outboundPluginTypes.jsOutbound),
  configuration: z.object({}),
});

export type JsOutboundFormSchemaType = z.infer<typeof JsOutboundFormSchema>;
