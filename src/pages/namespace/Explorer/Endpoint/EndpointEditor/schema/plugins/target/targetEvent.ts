import { targetPluginTypes } from ".";
import { z } from "zod";

export const TargetEventFormSchema = z.object({
  type: z.literal(targetPluginTypes.targetEvent),
  configuration: z.object({
    namespace: z.string().optional(),
  }),
});

export type TargetEventFormSchemaType = z.infer<typeof TargetEventFormSchema>;
