import { inboundPluginTypes } from ".";
import { z } from "zod";

export const EventFilterFormSchema = z.object({
  type: z.literal(inboundPluginTypes.eventFilter),
  configuration: z.object({
    script: z.string(),
    allow_non_events: z.boolean(),
  }),
});

export type EventFilterFormSchemaType = z.infer<typeof EventFilterFormSchema>;
