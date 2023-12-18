import { inboundPluginTypes } from ".";
import { z } from "zod";

export const JsInboundFormSchema = z.object({
  type: z.literal(inboundPluginTypes.jsInbound),
  configuration: z.object({
    script: z.string(),
  }),
});

export type JsInboundFormSchemaType = z.infer<typeof JsInboundFormSchema>;
