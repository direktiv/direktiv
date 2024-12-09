import { outboundPluginTypes } from ".";
import { z } from "zod";

export const JsOutboundFormSchema = z.object({
  type: z.literal(outboundPluginTypes.jsOutbound.name),
  configuration: z.object({
    script: z.string(),
  }),
});

export type JsOutboundFormSchemaType = z.infer<typeof JsOutboundFormSchema>;
